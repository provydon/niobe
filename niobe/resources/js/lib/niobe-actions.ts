export type StoredNiobeAction = {
    type: string;
    name: string;
    url?: string;
    target?: string;
};

export type NiobeAction = {
    type: string;
    name: string;
    target: string;
};

export type NiobeActionOption = {
    value: string;
    label: string;
    targetLabel: string;
    targetPlaceholder: string;
    hint: string;
};

export const niobeActionOptions: NiobeActionOption[] = [
    {
        value: 'send_email',
        label: 'Send email',
        targetLabel: 'Email address',
        targetPlaceholder: 'orders@example.com',
        hint: 'Use the inbox or recipient that should get the message.',
    },
    {
        value: 'send_webhook_event',
        label: 'Send webhook event',
        targetLabel: 'Webhook URL',
        targetPlaceholder: 'https://example.com/webhooks/niobe',
        hint: 'Niobe will send a payload to this endpoint.',
    },
    {
        value: 'send_whatsapp_message',
        label: 'Send WhatsApp message',
        targetLabel: 'WhatsApp number',
        targetPlaceholder: '+15551234567',
        hint: 'Use an international number or WhatsApp destination.',
    },
];

const legacyTypeMap: Record<string, string> = {
    order: 'send_webhook_event',
    ticket: 'send_webhook_event',
};

export function createEmptyNiobeAction(): NiobeAction {
    return {
        type: niobeActionOptions[0].value,
        name: 'Place order',
        target: '',
    };
}

export function getNiobeActionOption(type: string): NiobeActionOption {
    return niobeActionOptions.find((option) => option.value === type) ?? niobeActionOptions[0];
}

export function normalizeNiobeAction(
    action?: Partial<StoredNiobeAction> | null,
): NiobeAction {
    const normalizedType = legacyTypeMap[action?.type ?? ''] ?? action?.type ?? niobeActionOptions[0].value;

    return {
        type: normalizedType,
        name: action?.name ?? 'Place order',
        target: action?.target ?? action?.url ?? '',
    };
}
