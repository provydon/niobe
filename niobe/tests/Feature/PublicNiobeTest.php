<?php

use App\Models\User;
use Inertia\Testing\AssertableInertia as Assert;

test('public niobe page renders a talk link', function () {
    $user = User::factory()->create();
    $niobe = $user->niobes()->create([
        'name' => 'Cafe Helper',
        'slug' => 'cafe-helper',
        'context' => 'Help customers with menu questions and orders.',
        'tools' => [],
    ]);

    $this->get(route('niobe.show', $niobe->slug))
        ->assertOk()
        ->assertInertia(fn (Assert $page) => $page
            ->component('Niobe/Show')
            ->where('niobe.name', 'Cafe Helper')
            ->where('niobe.talk_url', route('niobe.talk', $niobe->slug))
        );
});

test('public niobe talk page renders websocket config', function () {
    config()->set('app.voice_agent_url', 'http://localhost:9000');

    $user = User::factory()->create();
    $niobe = $user->niobes()->create([
        'name' => 'Cafe Helper',
        'slug' => 'cafe-helper',
        'context' => 'Help customers with menu questions and orders.',
        'tools' => [],
    ]);

    $this->get(route('niobe.talk', $niobe->slug))
        ->assertOk()
        ->assertInertia(fn (Assert $page) => $page
            ->component('Niobe/Talk')
            ->where('niobe.name', 'Cafe Helper')
            ->where('voiceAgentWebsocketUrl', 'ws://localhost:9000/live?niobe=cafe-helper')
        );
});
