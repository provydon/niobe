<script setup lang="ts">
import { Head, Link, usePage } from '@inertiajs/vue3';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';

interface MenuItemRow {
    name: string;
    category: string;
    unit_price: string;
}

interface NiobeTalkProps {
    niobe: {
        name: string;
        context: string;
        menu?: string;
        share_url: string;
        menu_items?: MenuItemRow[];
        menu_image_urls?: string[];
        menu_currency?: string | null;
    };
    voiceAgentWebsocketUrl: string;
}

interface ToolEvent {
    callId: string;
    toolName: string;
    status: 'running' | 'success' | 'error';
    title: string;
    subtitle?: string;
    message?: string;
    actionType?: string;
}

interface IncomingToolEvent extends Omit<ToolEvent, 'status'> {
    status: 'running' | 'queued' | 'success' | 'error';
}

interface ConversationEvent {
    phase: string;
    status: string;
    tone: 'default' | 'connected' | 'error';
}

type TranscriptEntry =
    | { type: 'user'; id: number; text: string }
    | { type: 'agent'; id: number; text: string }
    | {
          type: 'action';
          id: number;
          callId: string;
          title: string;
          message?: string;
          status: 'running' | 'success' | 'error';
      };

const props = defineProps<NiobeTalkProps>();
let transcriptId = 0;

const page = usePage();
const appName = (page.props as { name?: string }).name ?? 'AI Waitress';

const talkTab = ref<'talk' | 'menu' | 'menu-images'>('talk');
const menuSearchQuery = ref('');
const menuPage = ref(1);
const menuPageSize = 10;

const allMenuItems = computed(() => props.niobe.menu_items ?? []);
const filteredMenuItems = computed(() => {
    const q = menuSearchQuery.value.trim().toLowerCase();
    if (!q) return allMenuItems.value;
    return allMenuItems.value.filter(
        (item) =>
            item.name.toLowerCase().includes(q) ||
            item.category.toLowerCase().includes(q) ||
            String(item.unit_price).toLowerCase().includes(q),
    );
});
const totalMenuPages = computed(() =>
    Math.max(1, Math.ceil(filteredMenuItems.value.length / menuPageSize)),
);
const paginatedMenuItems = computed(() => {
    const start = (menuPage.value - 1) * menuPageSize;
    return filteredMenuItems.value.slice(start, start + menuPageSize);
});

watch(menuSearchQuery, () => {
    menuPage.value = 1;
});
watch(talkTab, (tab) => {
    if (tab === 'menu') menuPage.value = 1;
});

const menuImageUrls = computed(() => props.niobe.menu_image_urls ?? []);
const zoomedMenuImageIndex = ref<number | null>(null);
const zoomOverlayRef = ref<HTMLElement | null>(null);

function openMenuImageZoom(idx: number) {
    zoomedMenuImageIndex.value = idx;
    nextTick(() => zoomOverlayRef.value?.focus());
}
function closeMenuImageZoom() {
    zoomedMenuImageIndex.value = null;
}

function onZoomKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') closeMenuImageZoom();
}

/** Format numeric price with thousands separator and 2 decimals (e.g. 1000 → "1,000.00"). Optionally append currency. */
function formatPrice(price: string | number, currency?: string | null): string {
    const n = Number(price);
    if (Number.isNaN(n)) return String(price);
    const formatted = new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
    }).format(n);
    if (currency && currency.trim()) return `${formatted} ${currency.trim()}`;
    return formatted;
}

const status = ref('Connecting...');
const statusTone = ref<'default' | 'connected' | 'error'>('default');
const isRecording = ref(false);
const isConnected = ref(false);
const microphoneSupported = ref(true);
const transcriptLines = ref<TranscriptEntry[]>([]);
const transcriptContainer = ref<HTMLElement | null>(null);
const preserveDisconnectStatus = ref(false);
let shouldAutoStart = false;

let ws: WebSocket | null = null;
let processor: ScriptProcessorNode | null = null;
let audioContext: AudioContext | null = null;
let source: MediaStreamAudioSourceNode | null = null;
let mediaStream: MediaStream | null = null;
const inputSampleRate = 16000;
const outputSampleRate = 24000;
const speechThreshold = 0.018;
const speechEndDelayMs = 550;
const maxBufferedFrames = 4;
const audio = new Audio();
const audioQueue: Uint8Array[] = [];
let isAudioPlaying = false;
let isSpeechActive = false;
let speechEndTimeout: number | null = null;
let bufferedFrames: Uint8Array[] = [];

type LegacyNavigator = Navigator & {
    getUserMedia?: (
        constraints: MediaStreamConstraints,
        successCallback: (stream: MediaStream) => void,
        errorCallback: (error: Error) => void
    ) => void;
    webkitGetUserMedia?: (
        constraints: MediaStreamConstraints,
        successCallback: (stream: MediaStream) => void,
        errorCallback: (error: Error) => void
    ) => void;
    mozGetUserMedia?: (
        constraints: MediaStreamConstraints,
        successCallback: (stream: MediaStream) => void,
        errorCallback: (error: Error) => void
    ) => void;
};

function setStatus(text: string, tone: 'default' | 'connected' | 'error' = 'default') {
    status.value = text;
    statusTone.value = tone;
}

function upsertToolEvent(event: IncomingToolEvent) {
    const status: 'running' | 'success' | 'error' =
        event.status === 'queued' ? 'success' : event.status;

    const index = transcriptLines.value.findIndex(
        (e) => e.type === 'action' && e.callId === event.callId,
    );

    if (index === -1) {
        transcriptLines.value.push({
            type: 'action',
            id: transcriptId++,
            callId: event.callId,
            title: event.title,
            message: event.message,
            status,
        });
    } else {
        const prev = transcriptLines.value[index];
        if (prev.type === 'action') {
            transcriptLines.value[index] = {
                ...prev,
                title: event.title ?? prev.title,
                message: event.message ?? prev.message,
                status,
            };
        }
    }
    scrollTranscriptToBottom();
}

function applyConversationEvent(event: ConversationEvent) {
    setStatus(event.status, event.tone);

    if (event.phase === 'completed') {
        preserveDisconnectStatus.value = true;
        window.setTimeout(() => {
            disconnect();
        }, 1800);
    }
}

function appendTranscript(role: 'user' | 'agent', text: string) {
    const trimmed = text.trim();
    if (!trimmed) return;
    const type = role === 'user' ? 'user' : 'agent';
    const lines = transcriptLines.value;
    const last = lines[lines.length - 1];

    if (last && last.type === type) {
        const sep = last.text.endsWith(' ') || trimmed.startsWith(' ') ? '' : ' ';
        transcriptLines.value = [
            ...lines.slice(0, -1),
            { ...last, text: last.text + sep + trimmed },
        ];
    } else {
        transcriptLines.value = [
            ...lines,
            { type, id: transcriptId++, text: trimmed } as TranscriptEntry,
        ];
    }
    scrollTranscriptToBottom();
}

function scrollTranscriptToBottom() {
    nextTick(() => {
        const el = transcriptContainer.value;
        if (el) el.scrollTop = el.scrollHeight;
    });
}

function mergeUint8(arrays: Uint8Array[]) {
    const total = arrays.reduce((acc, current) => acc + current.length, 0);
    const out = new Uint8Array(total);
    let offset = 0;

    for (const array of arrays) {
        out.set(array, offset);
        offset += array.length;
    }

    return out;
}

function encodeWav(chunks: Uint8Array[], rate: number, bitDepth: number, channels: number) {
    const data = mergeUint8(chunks);
    const dataSize = data.length;
    const fileSize = dataSize + 36;
    const blockAlign = (channels * bitDepth) / 8;
    const byteRate = rate * blockAlign;
    const buffer = new ArrayBuffer(44);
    const view = new DataView(buffer);
    const writeString = (offset: number, value: string) => {
        for (let i = 0; i < value.length; i += 1) {
            view.setUint8(offset + i, value.charCodeAt(i));
        }
    };

    writeString(0, 'RIFF');
    view.setUint32(4, fileSize, true);
    writeString(8, 'WAVE');
    writeString(12, 'fmt ');
    view.setUint32(16, 16, true);
    view.setUint16(20, 1, true);
    view.setUint16(22, channels, true);
    view.setUint32(24, rate, true);
    view.setUint32(28, byteRate, true);
    view.setUint16(32, blockAlign, true);
    view.setUint16(34, bitDepth, true);
    writeString(36, 'data');
    view.setUint32(40, dataSize, true);

    return new Blob([new Uint8Array(buffer), data], { type: 'audio/wav' });
}

function b64ToUint8(value: string) {
    const binary = atob(value);
    const out = new Uint8Array(binary.length);

    for (let i = 0; i < binary.length; i += 1) {
        out[i] = binary.charCodeAt(i);
    }

    return out;
}

/** Stop agent playback and clear the queue (e.g. on user interrupt or server interrupted). */
function stopAgentPlayback() {
    audioQueue.length = 0;
    if (isAudioPlaying && audio.src) {
        try {
            URL.revokeObjectURL(audio.src);
        } catch {
            /* ignore */
        }
        audio.pause();
        audio.removeAttribute('src');
        audio.load();
    }
    isAudioPlaying = false;
}

function playNext() {
    if (isAudioPlaying || audioQueue.length === 0) {
        return;
    }

    isAudioPlaying = true;
    const chunks = audioQueue.splice(0, audioQueue.length);
    const blob = encodeWav(chunks, outputSampleRate, 16, 1);
    audio.src = URL.createObjectURL(blob);
    audio.onended = () => {
        isAudioPlaying = false;
        URL.revokeObjectURL(audio.src);
        playNext();
    };
    audio.play().catch(() => {
        isAudioPlaying = false;
        playNext();
    });
}

function sendWsPayload(payload: Record<string, unknown>) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(payload));
    }
}

function clearPendingSpeechEnd() {
    if (speechEndTimeout !== null) {
        window.clearTimeout(speechEndTimeout);
        speechEndTimeout = null;
    }
}

function markSpeechStarted() {
    clearPendingSpeechEnd();

    if (isSpeechActive) {
        return;
    }

    isSpeechActive = true;
    // Stop agent playback so the user can talk naturally (barge-in).
    stopAgentPlayback();
}

function markSpeechEnded() {
    clearPendingSpeechEnd();

    if (!isSpeechActive) {
        return;
    }

    isSpeechActive = false;
    bufferedFrames = [];
    sendWsPayload({
        audioStreamEnd: true,
    });
}

function scheduleSpeechEnd() {
    clearPendingSpeechEnd();
    speechEndTimeout = window.setTimeout(() => {
        markSpeechEnded();
    }, speechEndDelayMs);
}

function bufferFrame(frame: Uint8Array) {
    bufferedFrames.push(frame);

    if (bufferedFrames.length > maxBufferedFrames) {
        bufferedFrames.shift();
    }
}

function sendAudioFrame(frame: Uint8Array) {
    const binary = Array.from(frame, (byte) => String.fromCharCode(byte)).join('');

    sendWsPayload({
        media: {
            data: btoa(binary),
            mimeType: 'audio/pcm;rate=16000',
        },
    });
}

function flushBufferedFrames() {
    for (const frame of bufferedFrames) {
        sendAudioFrame(frame);
    }

    bufferedFrames = [];
}

function handleSpeechFrame(samples: Float32Array, frame: Uint8Array) {
    let sumSquares = 0;

    for (let i = 0; i < samples.length; i += 1) {
        sumSquares += samples[i] * samples[i];
    }

    const rms = Math.sqrt(sumSquares / samples.length);

    if (rms >= speechThreshold) {
        if (!isSpeechActive) {
            flushBufferedFrames();
        }

        markSpeechStarted();
        sendAudioFrame(frame);

        return;
    }

    if (isSpeechActive) {
        sendAudioFrame(frame);
        scheduleSpeechEnd();

        return;
    }

    bufferFrame(frame);
}

function cleanupAudio() {
    markSpeechEnded();
    processor?.disconnect();
    source?.disconnect();
    mediaStream?.getTracks().forEach((track) => track.stop());
    void audioContext?.close();

    processor = null;
    source = null;
    mediaStream = null;
    audioContext = null;
    isRecording.value = false;
}

function recordStop() {
    cleanupAudio();
}

function hasGetUserMediaSupport() {
    const legacyNavigator = navigator as LegacyNavigator;

    return Boolean(
        navigator.mediaDevices?.getUserMedia
        || legacyNavigator.getUserMedia
        || legacyNavigator.webkitGetUserMedia
        || legacyNavigator.mozGetUserMedia,
    );
}

async function requestMicrophoneStream() {
    if (navigator.mediaDevices?.getUserMedia) {
        return navigator.mediaDevices.getUserMedia({ audio: true });
    }

    const legacyNavigator = navigator as LegacyNavigator;
    const legacyGetUserMedia =
        legacyNavigator.getUserMedia
        ?? legacyNavigator.webkitGetUserMedia
        ?? legacyNavigator.mozGetUserMedia;

    if (legacyGetUserMedia) {
        return new Promise<MediaStream>((resolve, reject) => {
            legacyGetUserMedia.call(legacyNavigator, { audio: true }, resolve, reject);
        });
    }

    throw new Error(
        window.isSecureContext
            ? 'This browser does not support microphone access on this page.'
            : 'Microphone access requires HTTPS or localhost.',
    );
}

async function recordStart() {
    if (isRecording.value) {
        return;
    }

    try {
        mediaStream = await requestMicrophoneStream();
        audioContext = new AudioContext({ sampleRate: inputSampleRate });
        source = audioContext.createMediaStreamSource(mediaStream);
        processor = audioContext.createScriptProcessor(1024, 1, 1);
        processor.onaudioprocess = (event) => {
            const float32 = event.inputBuffer.getChannelData(0);
            const int16 = new Int16Array(float32.length);

            for (let i = 0; i < float32.length; i += 1) {
                int16[i] = Math.max(-32768, Math.min(32767, float32[i] * 32768));
            }

            const bytes = new Uint8Array(int16.buffer);
            handleSpeechFrame(float32, bytes);
        };
        source.connect(processor);
        processor.connect(audioContext.destination);
        isRecording.value = true;
    } catch (error: any) {
        setStatus(`Microphone error: ${error.message}`, 'error');
    }
}

function toggleRecording() {
    if (isRecording.value) {
        recordStop();

        return;
    }

    void recordStart();
}

function disconnect() {
    recordStop();
    ws?.close();
}

function connect() {
    if (ws) {
        return;
    }

    setStatus('Connecting...');
    ws = new WebSocket(props.voiceAgentWebsocketUrl);
    ws.onopen = () => {
        isConnected.value = true;
        transcriptLines.value = [];

        if (microphoneSupported.value) {
            if (shouldAutoStart) {
                shouldAutoStart = false;
                setStatus('Connected. Starting microphone...', 'connected');
                void recordStart();

                return;
            }

            setStatus('Connected. Click “Start talking” and speak.', 'connected');

            return;
        }

        setStatus(
            window.isSecureContext
                ? 'Connected, but microphone is not supported in this browser.'
                : 'Connected, but microphone access requires HTTPS or localhost.',
            'error',
        );
    };
    ws.onclose = () => {
        ws = null;
        isConnected.value = false;
        recordStop();

        if (preserveDisconnectStatus.value) {
            preserveDisconnectStatus.value = false;

            return;
        }

        setStatus('Disconnected.');
    };
    ws.onerror = () => {
        setStatus('WebSocket error.', 'error');
    };
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);

        if (data.error) {
            setStatus(`Error: ${data.error}`, 'error');

            return;
        }

        if (data.toolEvent) {
            upsertToolEvent(data.toolEvent);
        }

        if (data.conversationEvent) {
            applyConversationEvent(data.conversationEvent);
        }

        if (data.transcript?.role === 'agent' && data.transcript?.text) {
            appendTranscript('agent', data.transcript.text);
        }

        if (data.serverContent?.interrupted) {
            stopAgentPlayback();
        }

        if (!data.serverContent) {
            return;
        }

        const turn = data.serverContent.modelTurn;
        if (turn?.parts) {
            for (const part of turn.parts) {
                if (
                    part.inlineData &&
                    part.inlineData.mimeType &&
                    part.inlineData.mimeType.startsWith('audio') &&
                    part.inlineData.data
                ) {
                    audioQueue.push(b64ToUint8(part.inlineData.data));
                }
            }
        }

        if (data.serverContent.turnComplete) {
            if (audioQueue.length > 0) {
                playNext();
            }
        } else if (audioQueue.length > 0) {
            playNext();
        }
    };
}

onMounted(() => {
    microphoneSupported.value = hasGetUserMediaSupport();
    shouldAutoStart = new URLSearchParams(window.location.search).get('autostart') === '1';

    if (shouldAutoStart) {
        const url = new URL(window.location.href);
        url.searchParams.delete('autostart');
        window.history.replaceState({}, '', `${url.pathname}${url.search}${url.hash}`);
    }

    if (!microphoneSupported.value) {
        setStatus(
            window.isSecureContext
                ? 'Microphone is not supported in this browser.'
                : 'Microphone access requires HTTPS or localhost.',
            'error',
        );
    }

    connect();
});
onBeforeUnmount(disconnect);
</script>

<template>
    <div class="flex min-h-screen flex-col items-center bg-[#0f0f12] px-4 pt-6 pb-10 text-[#e4e4e7] sm:px-6">
        <Head :title="`${niobe.name} - ${appName}`" />

        <div class="w-full max-w-[32rem] min-h-[32rem] rounded-2xl border border-white/10 bg-[#17171c] p-6 shadow-2xl sm:p-8 flex flex-col">
            <div class="mb-4 flex shrink-0 border-b border-white/10">
                <button
                    type="button"
                    class="border-b-2 px-4 py-3 text-sm font-medium transition-colors sm:px-6"
                    :class="talkTab === 'talk' ? '-mb-px border-[#3b82f6] text-white' : 'border-transparent text-[#71717a] hover:text-[#e4e4e7]'"
                    @click="talkTab = 'talk'"
                >
                    Talk
                </button>
                <button
                    type="button"
                    class="border-b-2 px-4 py-3 text-sm font-medium transition-colors sm:px-6"
                    :class="talkTab === 'menu' ? '-mb-px border-[#3b82f6] text-white' : 'border-transparent text-[#71717a] hover:text-[#e4e4e7]'"
                    @click="talkTab = 'menu'"
                >
                    Menu
                </button>
                <button
                    type="button"
                    class="border-b-2 px-4 py-3 text-sm font-medium transition-colors sm:px-6"
                    :class="talkTab === 'menu-images' ? '-mb-px border-[#3b82f6] text-white' : 'border-transparent text-[#71717a] hover:text-[#e4e4e7]'"
                    @click="talkTab = 'menu-images'"
                >
                    Menu images
                </button>
            </div>

            <div class="min-h-[24rem] flex-1 overflow-y-auto">
            <div v-show="talkTab === 'talk'" class="space-y-4">
            <div class="mb-6">
                <p class="text-xs uppercase tracking-[0.2em] text-[#71717a]">AI waitress</p>
                <h1 class="mt-2 text-2xl font-semibold">{{ niobe.name }}</h1>
                <p
                    class="mt-3 text-sm"
                    :class="{
                        'text-[#71717a]': statusTone === 'default',
                        'text-[#22c55e]': statusTone === 'connected',
                        'text-[#ef4444]': statusTone === 'error',
                    }"
                >
                    {{ status }}
                </p>
            </div>

            <div class="space-y-3">
                <button
                    type="button"
                    class="w-full rounded-xl px-5 py-3 text-sm font-medium text-white transition disabled:cursor-not-allowed disabled:opacity-60"
                    :class="isRecording ? 'bg-[#dc2626] hover:bg-[#b91c1c]' : 'bg-[#3b82f6] hover:bg-[#2563eb]'"
                    :disabled="!isConnected || !microphoneSupported"
                    @click="toggleRecording"
                >
                    {{ isRecording ? 'Stop' : 'Start talking' }}
                </button>

                <button
                    type="button"
                    class="w-full rounded-xl bg-[#27272a] px-5 py-3 text-sm font-medium text-[#a1a1aa] transition hover:bg-[#3f3f46]"
                    @click="disconnect"
                >
                    Disconnect
                </button>
            </div>

            <div class="mt-6 space-y-3">
                <div class="flex items-center justify-between">
                    <p class="text-xs uppercase tracking-[0.2em] text-[#71717a]">Live chat</p>
                </div>
                <div
                    ref="transcriptContainer"
                    class="max-h-72 min-h-[8rem] overflow-y-auto rounded-xl border border-white/10 bg-[#111115] px-4 py-3 text-sm"
                >
                    <div v-if="transcriptLines.length" class="space-y-3">
                        <div
                            v-for="entry in transcriptLines"
                            :key="entry.id"
                            class="flex"
                            :class="{
                                'justify-end': entry.type === 'user',
                                'justify-start': entry.type === 'agent',
                                'justify-center': entry.type === 'action',
                            }"
                        >
                            <!-- User: interpreted text sent to the model -->
                            <p
                                v-if="entry.type === 'user'"
                                class="max-w-[85%] rounded-lg bg-[#3b82f6] px-3 py-2 text-white"
                            >
                                <span class="mr-2 text-blue-200">You (sent to model):</span>
                                {{ entry.text }}
                            </p>
                            <!-- Agent -->
                            <p
                                v-else-if="entry.type === 'agent'"
                                class="max-w-[85%] rounded-lg bg-[#27272a] px-3 py-2 text-[#e4e4e7]"
                            >
                                <span class="mr-2 text-[#71717a]">Waitress:</span>
                                {{ entry.text }}
                            </p>
                            <!-- Action (tool run) – clearly differentiated from chat -->
                            <div
                                v-else-if="entry.type === 'action'"
                                class="flex w-full max-w-[90%] overflow-hidden rounded-xl border-l-4 text-sm shadow-sm"
                                :class="{
                                    'border-l-[#3b82f6] bg-[#0f172a] border border-white/10': entry.status === 'running',
                                    'border-l-[#22c55e] bg-[#052e16]/80 border border-[#14532d]/60': entry.status === 'success',
                                    'border-l-[#ef4444] bg-[#450a0a]/80 border border-[#7f1d1d]/60': entry.status === 'error',
                                }"
                            >
                                <div class="flex flex-1 flex-col gap-1 px-4 py-3">
                                    <div class="flex items-center gap-2">
                                        <span
                                            class="rounded bg-white/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-[#94a3b8]"
                                        >
                                            Action
                                        </span>
                                        <span
                                            class="shrink-0 text-lg leading-none"
                                            :class="{
                                                'text-[#93c5fd]': entry.status === 'running',
                                                'text-[#22c55e]': entry.status === 'success',
                                                'text-[#ef4444]': entry.status === 'error',
                                            }"
                                            :aria-label="entry.status"
                                        >
                                            {{ entry.status === 'running' ? '⏳' : entry.status === 'success' ? '✓' : '✗' }}
                                        </span>
                                        <span class="font-semibold text-white">{{ entry.title }}</span>
                                        <span
                                            class="ml-auto rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider"
                                            :class="{
                                                'bg-[#1e3a5f] text-[#93c5fd]': entry.status === 'running',
                                                'bg-[#14532d] text-[#86efac]': entry.status === 'success',
                                                'bg-[#7f1d1d] text-[#fecaca]': entry.status === 'error',
                                            }"
                                        >
                                            {{ entry.status }}
                                        </span>
                                    </div>
                                    <p v-if="entry.message" class="text-xs text-[#cbd5e1]">
                                        {{ entry.message }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                    <p v-else class="text-[#71717a]">
                        Replies and actions will appear here.
                    </p>
                </div>
            </div>

            <p class="mt-5 text-sm leading-relaxed text-[#a1a1aa]">
                Allow microphone, then talk. Your waitress replies with voice.
            </p>

            <div class="mt-8 flex flex-wrap items-center gap-4 text-sm">
                <Link :href="niobe.share_url" class="text-[#93c5fd] hover:underline">
                    Back
                </Link>
                <span class="text-[#52525b]">·</span>
                <span class="text-[#71717a]">{{ appName }}</span>
            </div>
            </div>

            <div v-show="talkTab === 'menu'" class="space-y-4">
                <p class="text-xs uppercase tracking-[0.2em] text-[#71717a]">Menu</p>
                <div v-if="allMenuItems.length > 0" class="mt-3 space-y-3">
                    <input
                        v-model="menuSearchQuery"
                        type="search"
                        placeholder="Search menu..."
                        class="w-full rounded-lg border border-white/10 bg-[#111115] px-3 py-2 text-sm text-[#e4e4e7] placeholder:text-[#71717a] focus:border-[#3b82f6] focus:outline-none focus:ring-1 focus:ring-[#3b82f6]"
                    />
                    <div class="overflow-hidden rounded-xl border border-white/10 bg-[#111115]">
                        <div class="grid grid-cols-1 gap-0 sm:grid-cols-[auto_1fr_auto]">
                            <div class="border-b border-white/10 bg-[#0f0f12] px-3 py-2 text-xs font-semibold uppercase tracking-wider text-[#71717a] sm:border-b-0 sm:border-r">Name</div>
                            <div class="border-b border-white/10 bg-[#0f0f12] px-3 py-2 text-xs font-semibold uppercase tracking-wider text-[#71717a] sm:border-b-0 sm:border-r">Category</div>
                            <div class="border-b border-white/10 bg-[#0f0f12] px-3 py-2 text-xs font-semibold uppercase tracking-wider text-[#71717a]">Price</div>
                            <template v-if="paginatedMenuItems.length">
                            <template v-for="(item, i) in paginatedMenuItems" :key="i">
                                <div class="border-b border-white/5 px-3 py-2 text-sm text-[#e4e4e7] last:border-b-0 sm:border-b-0 sm:border-r">{{ item.name }}</div>
                                <div class="border-b border-white/5 px-3 py-2 text-sm text-[#a1a1aa] last:border-b-0 sm:border-b-0 sm:border-r">{{ item.category }}</div>
                                <div class="border-b border-white/5 px-3 py-2 text-sm text-[#e4e4e7] last:border-b-0">{{ formatPrice(item.unit_price, niobe.menu_currency) }}</div>
                            </template>
                            </template>
                            <template v-else>
                                <div class="col-span-3 px-3 py-4 text-center text-sm text-[#71717a]">
                                    No items match your search.
                                </div>
                            </template>
                        </div>
                    </div>
                    <div v-if="totalMenuPages > 1" class="flex flex-wrap items-center justify-center gap-2">
                        <button
                            type="button"
                            class="rounded-lg border border-white/10 bg-[#27272a] px-3 py-1.5 text-sm text-[#e4e4e7] hover:bg-[#3f3f46] disabled:opacity-50"
                            :disabled="menuPage <= 1"
                            @click="menuPage = Math.max(1, menuPage - 1)"
                        >
                            Previous
                        </button>
                        <span class="text-sm text-[#71717a]">
                            Page {{ menuPage }} of {{ totalMenuPages }}
                        </span>
                        <button
                            type="button"
                            class="rounded-lg border border-white/10 bg-[#27272a] px-3 py-1.5 text-sm text-[#e4e4e7] hover:bg-[#3f3f46] disabled:opacity-50"
                            :disabled="menuPage >= totalMenuPages"
                            @click="menuPage = Math.min(totalMenuPages, menuPage + 1)"
                        >
                            Next
                        </button>
                    </div>
                </div>
                <div v-else class="mt-4 rounded-xl border border-white/10 bg-[#111115] px-4 py-3 text-sm text-[#e4e4e7] whitespace-pre-wrap">{{ niobe.menu ?? niobe.context ?? 'No menu set.' }}</div>
                <div class="mt-6 flex flex-wrap items-center gap-4 text-sm">
                    <Link :href="niobe.share_url" class="text-[#93c5fd] hover:underline">Back</Link>
                    <span class="text-[#52525b]">·</span>
                    <span class="text-[#71717a]">{{ appName }}</span>
                </div>
            </div>

            <div v-show="talkTab === 'menu-images'" class="space-y-4">
                <p class="text-xs uppercase tracking-[0.2em] text-[#71717a]">Menu images</p>
                <div v-if="menuImageUrls.length > 0" class="mt-3 space-y-8">
                    <div
                        v-for="(url, idx) in menuImageUrls"
                        :key="idx"
                        class="overflow-hidden rounded-xl border border-white/10 bg-[#111115]"
                    >
                        <p class="px-3 py-2 text-xs font-medium text-[#71717a]">Image {{ idx + 1 }} of {{ menuImageUrls.length }}</p>
                        <button
                            type="button"
                            class="block w-full cursor-zoom-in text-left"
                            @click="openMenuImageZoom(idx)"
                        >
                            <img
                                :src="url"
                                :alt="`Menu image ${idx + 1}`"
                                class="max-w-full align-top"
                            />
                        </button>
                    </div>
                    <!-- Zoom overlay -->
                    <Teleport to="body">
                        <div
                            ref="zoomOverlayRef"
                            v-if="zoomedMenuImageIndex !== null && menuImageUrls[zoomedMenuImageIndex]"
                            class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-4"
                            role="dialog"
                            aria-modal="true"
                            aria-label="Menu image zoomed"
                            tabindex="-1"
                            @click.self="closeMenuImageZoom"
                            @keydown="onZoomKeydown"
                        >
                            <button
                                type="button"
                                class="absolute right-4 top-4 rounded-full bg-white/10 p-2 text-white hover:bg-white/20"
                                aria-label="Close"
                                @click="closeMenuImageZoom"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M18 6 6 18"/><path d="m6 6 12 12"/>
                                </svg>
                            </button>
                            <img
                                :src="menuImageUrls[zoomedMenuImageIndex]"
                                :alt="`Menu image ${zoomedMenuImageIndex + 1}`"
                                class="max-h-[90vh] max-w-full object-contain"
                                @click.stop
                            />
                        </div>
                    </Teleport>
                </div>
                <div v-else class="mt-4 rounded-xl border border-white/10 bg-[#111115] px-4 py-8 text-center text-sm text-[#71717a]">
                    No menu images available.
                </div>
                <div class="mt-6 flex flex-wrap items-center gap-4 text-sm">
                    <Link :href="niobe.share_url" class="text-[#93c5fd] hover:underline">Back</Link>
                    <span class="text-[#52525b]">·</span>
                    <span class="text-[#71717a]">{{ appName }}</span>
                </div>
            </div>
            </div>
        </div>
    </div>
</template>
