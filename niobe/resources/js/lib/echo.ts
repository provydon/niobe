import Echo from 'laravel-echo';
import Pusher from 'pusher-js';

const DEBUG_PREFIX = '[Echo]';

declare global {
    interface Window {
        Echo?: InstanceType<typeof Echo>;
        Pusher?: typeof Pusher;
    }
}

const key = import.meta.env.VITE_PUSHER_APP_KEY as string | undefined;
const cluster = (import.meta.env.VITE_PUSHER_APP_CLUSTER as string) || 'mt1';

export function isBroadcastingEnabled(): boolean {
    return !!key;
}

export function getEcho(): Echo | null {
    if (!key) {
        console.debug(DEBUG_PREFIX, 'Broadcasting disabled (no VITE_PUSHER_APP_KEY)');
        return null;
    }
    if (!window.Echo) {
        console.debug(DEBUG_PREFIX, 'Initializing Echo', { key: key.slice(0, 8) + '…', cluster });
        window.Pusher = Pusher;
        window.Echo = new Echo({
            broadcaster: 'pusher',
            key,
            cluster,
            forceTLS: true,
        });
        const pusher = (window.Echo as unknown as { connector: { pusher: { connection: { bind: (event: string, cb: () => void) => void } } } }).connector?.pusher;
        if (pusher?.connection) {
            pusher.connection.bind('connecting', () => console.debug(DEBUG_PREFIX, 'Connection: connecting…'));
            pusher.connection.bind('connected', () => console.debug(DEBUG_PREFIX, 'Connection: established'));
            pusher.connection.bind('disconnected', () => console.debug(DEBUG_PREFIX, 'Connection: disconnected'));
            pusher.connection.bind('failed', () => console.debug(DEBUG_PREFIX, 'Connection: failed'));
            pusher.connection.bind('unavailable', () => console.debug(DEBUG_PREFIX, 'Connection: unavailable'));
        }
    }
    return window.Echo;
}

export type DataRecordUpdatedPayload = {
    dataId: number;
    status: string;
    batchesDone?: number;
    batchesTotal?: number;
};

/** Subscribe to real-time updates for a data record. Returns unsubscribe function. */
export function subscribeDataRecord(
    dataId: number,
    onUpdate: (payload: DataRecordUpdatedPayload) => void
): () => void {
    const echo = getEcho();
    if (!echo) return () => {};
    const channelName = `data.${dataId}`;
    console.debug(DEBUG_PREFIX, 'Subscribing to channel', channelName);
    const channel = echo.channel(channelName);
    channel.listen('.DataRecordUpdated', (e: DataRecordUpdatedPayload) => {
        console.debug(DEBUG_PREFIX, 'Received DataRecordUpdated', { channel: channelName, payload: e });
        onUpdate(e);
    });
    const pusher = (echo as unknown as { connector: { pusher: Pusher } }).connector?.pusher;
    if (pusher) {
        const bindSubscriptionSucceeded = () => {
            const ch = pusher.channel(channelName);
            if (ch?.bind) {
                ch.bind('pusher:subscription_succeeded', () => console.debug(DEBUG_PREFIX, 'Subscription succeeded', channelName));
            }
        };
        if (pusher.channel(channelName)) bindSubscriptionSucceeded();
        else setTimeout(bindSubscriptionSucceeded, 100);
    }
    return () => {
        console.debug(DEBUG_PREFIX, 'Leaving channel', channelName);
        echo.leave(channelName);
    };
}
