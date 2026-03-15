<?php

use App\Jobs\ExtractMenuJob;
use App\Models\User;
use App\Models\Waitress;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Queue;
use Illuminate\Support\Facades\Storage;

function validActions(): array
{
    return [
        [
            'type' => 'send_email',
            'name' => 'When a customer asks for a receipt',
            'target' => 'receipts@example.com',
        ],
        [
            'type' => 'send_webhook_event',
            'name' => 'When a customer places an order',
            'target' => 'https://example.com/webhooks/orders',
        ],
    ];
}

test('authenticated users can create a waitress with actions', function () {
    Storage::fake('local');
    Queue::fake();

    $user = User::factory()->create();
    $file = UploadedFile::fake()->image('menu.jpg', 100, 100);

    $response = $this
        ->actingAs($user)
        ->post(route('waitresses.store'), [
            'name' => 'Cafe Helper',
            'actions' => validActions(),
            'menu_files' => [$file],
        ]);

    $waitress = Waitress::firstOrFail();

    $response->assertRedirect(route('waitresses.edit', $waitress, absolute: false));

    expect($waitress->user_id)->toBe($user->id)
        ->and($waitress->name)->toBe('Cafe Helper')
        ->and($waitress->context)->toBe('')
        ->and($waitress->tools)->toBe([
            [
                'type' => 'send_email',
                'name' => 'When a customer asks for a receipt',
                'target' => 'receipts@example.com',
            ],
            [
                'type' => 'send_webhook_event',
                'name' => 'When a customer places an order',
                'target' => 'https://example.com/webhooks/orders',
            ],
        ]);

    Queue::assertPushed(ExtractMenuJob::class, fn (ExtractMenuJob $job) => $job->waitressId === $waitress->id);
});

test('api create returns json instead of redirect', function () {
    Storage::fake('local');
    Queue::fake();

    $user = User::factory()->create();
    $file = UploadedFile::fake()->image('menu.jpg', 100, 100);

    $response = $this
        ->actingAs($user)
        ->postJson(route('waitresses.store'), [
            'name' => 'Cafe Helper',
            'actions' => validActions(),
            'menu_files' => [$file],
        ]);

    $waitress = Waitress::firstOrFail();

    $response->assertCreated()
        ->assertJsonPath('message', 'Waitress created. Menu is being extracted and will appear shortly.')
        ->assertJsonPath('redirect', route('waitresses.index'))
        ->assertJsonPath('waitress.id', $waitress->id);
});

test('creating a waitress with menu files queues ExtractMenuJob', function () {
    Storage::fake('local');
    Queue::fake();

    $user = User::factory()->create();
    $file = UploadedFile::fake()->image('menu.png', 200, 200);

    $response = $this
        ->actingAs($user)
        ->post(route('waitresses.store'), [
            'name' => 'Cafe Helper',
            'actions' => validActions(),
            'menu_files' => [$file],
        ]);

    $waitress = Waitress::firstOrFail();

    $response->assertRedirect(route('waitresses.edit', $waitress, absolute: false));
    $response->assertSessionHas('success', 'Waitress created. Menu is being extracted and will appear shortly.');

    Queue::assertPushed(ExtractMenuJob::class, function (ExtractMenuJob $job) use ($waitress) {
        expect($job->waitressId)->toBe($waitress->id)
            ->and($job->storagePaths)->toHaveCount(1)
            ->and(str_contains($job->storagePaths[0], 'waitresses/'))
            ->toBeTrue();

        return true;
    });
});

test('creating a waitress requires menu files', function () {
    $user = User::factory()->create();

    $response = $this
        ->actingAs($user)
        ->from(route('waitresses.create'))
        ->post(route('waitresses.store'), [
            'name' => 'Cafe Helper',
            'actions' => validActions(),
        ]);

    $response->assertRedirect(route('waitresses.create', absolute: false));
    $response->assertSessionHasErrors(['menu_files']);

    expect(Waitress::count())->toBe(0);
});

test('creating a waitress requires at least one action', function () {
    $user = User::factory()->create();
    $file = UploadedFile::fake()->image('menu.jpg', 100, 100);

    $response = $this
        ->actingAs($user)
        ->from(route('waitresses.create'))
        ->post(route('waitresses.store'), [
            'name' => 'Cafe Helper',
            'actions' => [],
            'menu_files' => [$file],
        ]);

    $response->assertRedirect(route('waitresses.create', absolute: false));
    $response->assertSessionHasErrors(['actions']);

    expect(Waitress::count())->toBe(0);
});

test('creating a waitress rejects unsupported action types', function () {
    $user = User::factory()->create();
    $file = UploadedFile::fake()->image('menu.jpg', 100, 100);

    $response = $this
        ->actingAs($user)
        ->from(route('waitresses.create'))
        ->post(route('waitresses.store'), [
            'name' => 'Cafe Helper',
            'actions' => [
                [
                    'type' => 'launch_rocket',
                    'name' => 'When a customer says hello',
                    'target' => 'mission-control',
                ],
            ],
            'menu_files' => [$file],
        ]);

    $response->assertRedirect(route('waitresses.create', absolute: false));
    $response->assertSessionHasErrors(['actions.0.type']);

    expect(Waitress::count())->toBe(0);
});

test('authenticated users can update a waitress with new actions', function () {
    $user = User::factory()->create();
    $waitress = $user->waitresses()->create([
        'name' => 'Cafe Helper',
        'slug' => 'cafe-helper',
        'context' => 'Original context',
        'tools' => [
            [
                'type' => 'send_webhook_event',
                'name' => 'When an order is placed',
                'target' => 'https://example.com/webhooks/orders',
            ],
        ],
    ]);

    $response = $this
        ->actingAs($user)
        ->from(route('waitresses.edit', $waitress))
        ->patch(route('waitresses.update', $waitress), [
            'name' => 'Cafe Concierge',
            'actions' => [
                [
                    'type' => 'send_whatsapp_message',
                    'name' => 'When VIP customers ask for support',
                    'target' => '+15551234567',
                ],
                [
                    'type' => 'send_webhook_event',
                    'name' => 'When a delivery is delayed',
                    'target' => 'https://example.com/webhooks/delivery',
                ],
            ],
        ]);

    $response->assertRedirect(route('waitresses.edit', $waitress, absolute: false));

    $waitress->refresh();

    expect($waitress->name)->toBe('Cafe Concierge')
        ->and($waitress->tools)->toBe([
            [
                'type' => 'send_whatsapp_message',
                'name' => 'When VIP customers ask for support',
                'target' => '+15551234567',
            ],
            [
                'type' => 'send_webhook_event',
                'name' => 'When a delivery is delayed',
                'target' => 'https://example.com/webhooks/delivery',
            ],
        ]);
});
