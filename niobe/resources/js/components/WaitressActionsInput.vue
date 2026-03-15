<script setup lang="ts">
import { computed } from 'vue';
import InputError from '@/components/InputError.vue';
import { type NiobeAction, type NiobeActionOption } from '@/lib/niobe-actions';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

const props = withDefaults(
    defineProps<{
        modelValue: NiobeAction[];
        actionTypes: NiobeActionOption[];
        errors?: Record<string, string>;
        description?: string;
        variant?: 'create' | 'edit';
    }>(),
    {
        description: 'Where to send orders (e.g. email or webhook)',
        errors: () => ({}),
        variant: 'edit',
    },
);

const emit = defineEmits<{
    'update:modelValue': [value: NiobeAction[]];
}>();

const actions = computed({
    get: () => props.modelValue,
    set: (value: NiobeAction[]) => emit('update:modelValue', value),
});

function emptyAction(type?: string): NiobeAction {
    const first = props.actionTypes[0]?.value ?? 'send_email';
    return { type: type ?? first, name: 'Place order', target: '' };
}

function getActionOption(type: string): NiobeActionOption {
    return props.actionTypes.find((o) => o.value === type) ?? props.actionTypes[0]!;
}

function addAction() {
    emit('update:modelValue', [...props.modelValue, emptyAction()]);
}

function removeAction(i: number) {
    if (props.modelValue.length <= 1) return;
    emit('update:modelValue', props.modelValue.filter((_, idx) => idx !== i));
}

const borderClass = computed(() =>
    props.variant === 'create'
        ? 'rounded-md border-2 border-border bg-muted/30 p-3'
        : 'rounded-md border border-border bg-muted/50 p-3',
);
const selectClass = computed(() =>
    props.variant === 'create'
        ? 'h-9 w-full rounded-md border-2 border-border bg-background px-3 text-sm'
        : 'h-9 w-full rounded-md border border-input bg-background px-3 text-sm',
);
const inputClass = computed(() => (props.variant === 'create' ? 'border-2 border-border' : ''));
</script>

<template>
    <div class="space-y-2">
        <div class="space-y-1">
            <Label>Actions</Label>
            <p class="text-xs text-muted-foreground">
                {{ description }}
            </p>
        </div>
        <div
            v-for="(action, i) in actions"
            :key="i"
            class="space-y-3"
            :class="borderClass"
        >
            <input :name="`actions[${i}][name]`" type="hidden" :value="action.name || 'Place order'" />
            <div class="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,200px)_minmax(0,1fr)_auto] md:items-start">
                <div class="space-y-1">
                    <Label :for="`action-type-${i}`" class="text-xs text-muted-foreground">
                        {{ variant === 'create' ? 'Type' : 'Action' }}
                    </Label>
                    <select
                        :id="`action-type-${i}`"
                        :name="`actions[${i}][type]`"
                        v-model="action.type"
                        :class="selectClass"
                    >
                        <option
                            v-for="option in actionTypes"
                            :key="option.value"
                            :value="option.value"
                        >
                            {{ option.label }}
                        </option>
                    </select>
                    <InputError :message="errors[`actions.${i}.type`]" />
                </div>
                <div class="min-w-0 space-y-1">
                    <Label :for="`action-target-${i}`" class="text-xs text-muted-foreground">
                        {{ getActionOption(action.type).targetLabel }}
                    </Label>
                    <Input
                        :id="`action-target-${i}`"
                        :name="`actions[${i}][target]`"
                        v-model="action.target"
                        :type="action.type === 'send_email' ? 'email' : action.type === 'send_webhook_event' ? 'url' : 'text'"
                        :placeholder="getActionOption(action.type).targetPlaceholder"
                        :class="inputClass"
                    />
                    <InputError :message="errors[`actions.${i}.target`]" />
                </div>
                <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    class="text-destructive"
                    :disabled="modelValue.length === 1"
                    @click="removeAction(i)"
                >
                    Remove
                </Button>
            </div>
            <p class="text-xs text-muted-foreground">
                {{ getActionOption(action.type).hint }}
            </p>
        </div>
        <div class="flex justify-start">
            <Button type="button" variant="secondary" size="sm" class="border border-border" @click="addAction">
                Add another action
            </Button>
        </div>
        <p v-if="variant === 'edit'" class="text-xs text-muted-foreground">At least one action is required.</p>
        <InputError :message="errors.actions" />
    </div>
</template>
