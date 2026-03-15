import { onBeforeUnmount, ref } from 'vue';

interface SpeechRecognitionAlternativeLike {
    transcript: string;
}

interface SpeechRecognitionResultLike {
    0: SpeechRecognitionAlternativeLike;
    isFinal: boolean;
    length: number;
}

interface SpeechRecognitionEventLike {
    resultIndex: number;
    results: ArrayLike<SpeechRecognitionResultLike>;
}

interface SpeechRecognitionErrorEventLike {
    error: string;
}

interface BrowserSpeechRecognition {
    continuous: boolean;
    interimResults: boolean;
    lang: string;
    onresult: ((event: SpeechRecognitionEventLike) => void) | null;
    onerror: ((event: SpeechRecognitionErrorEventLike) => void) | null;
    onend: (() => void) | null;
    start: () => void;
    stop: () => void;
}

type SpeechRecognitionConstructor = new () => BrowserSpeechRecognition;

type SpeechRecognitionWindow = Window & {
    SpeechRecognition?: SpeechRecognitionConstructor;
    webkitSpeechRecognition?: SpeechRecognitionConstructor;
};

function getRecognitionConstructor(): SpeechRecognitionConstructor | null {
    if (typeof window === 'undefined') {
        return null;
    }

    const browserWindow = window as SpeechRecognitionWindow;

    return browserWindow.SpeechRecognition ?? browserWindow.webkitSpeechRecognition ?? null;
}

export function useVoiceInput(onTranscript: (text: string) => void) {
    const isListening = ref(false);
    const voiceError = ref<string | null>(null);
    const isVoiceInputSupported = Boolean(getRecognitionConstructor());

    let recognition: BrowserSpeechRecognition | null = null;

    const stopListening = () => {
        recognition?.stop();
    };

    const startListening = () => {
        const Recognition = getRecognitionConstructor();

        if (!Recognition) {
            voiceError.value = 'Voice input is not supported in this browser.';
            return;
        }

        if (isListening.value) {
            return;
        }

        voiceError.value = null;
        recognition = new Recognition();
        recognition.continuous = true;
        recognition.interimResults = false;
        recognition.lang = 'en-US';

        recognition.onresult = (event) => {
            const transcriptParts: string[] = [];

            for (let i = event.resultIndex; i < event.results.length; i += 1) {
                const result = event.results[i];

                if (result.isFinal && result[0]?.transcript) {
                    transcriptParts.push(result[0].transcript.trim());
                }
            }

            const transcript = transcriptParts.join(' ').trim();

            if (transcript) {
                onTranscript(transcript);
            }
        };

        recognition.onerror = (event) => {
            voiceError.value = event.error === 'not-allowed'
                ? 'Microphone access was denied.'
                : 'Voice input failed. Please try again.';
            isListening.value = false;
            recognition = null;
        };

        recognition.onend = () => {
            isListening.value = false;
            recognition = null;
        };

        recognition.start();
        isListening.value = true;
    };

    const toggleListening = () => {
        if (isListening.value) {
            stopListening();
            return;
        }

        startListening();
    };

    onBeforeUnmount(stopListening);

    return {
        isListening,
        isVoiceInputSupported,
        startListening,
        stopListening,
        toggleListening,
        voiceError,
    };
}
