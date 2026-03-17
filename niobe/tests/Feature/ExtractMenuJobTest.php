<?php

use App\Jobs\ExtractMenuJob;
use App\Models\User;
use App\Services\MenuExtractionService;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Storage;

beforeEach(function () {
    Storage::fake();
});

test('job skips empty paths and does not call storage with empty key', function () {
    $user = User::factory()->create();
    $waitress = $user->waitresses()->create([
        'name' => 'Test',
        'slug' => 'test',
        'context' => '',
        'tools' => [['type' => 'send_email', 'name' => 'n', 'target' => 't@t.com']],
    ]);
    $path = "waitresses/{$waitress->id}/menu-extraction/menu.jpg";
    Storage::put($path, UploadedFile::fake()->image('menu.jpg', 10, 10)->get());

    $extractor = $this->mock(MenuExtractionService::class);
    $extractor->shouldReceive('extractMenuFromFiles')->once()->andReturn([
        'currency' => 'USD',
        'items' => [['name' => 'Coffee', 'category' => 'Drinks', 'unit_price' => 3.50]],
    ]);

    $job = new ExtractMenuJob($waitress->id, ['', '  ', $path]);
    $job->handle($extractor);

    $waitress->refresh();
    expect($waitress->menuItems)->toHaveCount(1)
        ->and($waitress->menuItems->first()->name)->toBe('Coffee');
});

test('job deletes stored files when waitress is missing', function () {
    $path = 'waitresses/99999/menu-extraction/orphan.jpg';
    Storage::put($path, 'content');

    $job = new ExtractMenuJob(99999, [$path]);
    $job->handle(app(MenuExtractionService::class));

    expect(Storage::exists($path))->toBeFalse();
});

test('job runs extraction and updates waitress when extraction returns no items', function () {
    $user = User::factory()->create();
    $waitress = $user->waitresses()->create([
        'name' => 'Test',
        'slug' => 'test',
        'context' => '',
        'tools' => [['type' => 'send_email', 'name' => 'n', 'target' => 't@t.com']],
    ]);
    $path = "waitresses/{$waitress->id}/menu-extraction/menu.jpg";
    Storage::put($path, UploadedFile::fake()->image('menu.jpg', 10, 10)->get());

    $extractor = $this->mock(MenuExtractionService::class);
    $extractor->shouldReceive('extractMenuFromFiles')->once()->andReturn(['currency' => 'USD', 'items' => []]);

    $job = new ExtractMenuJob($waitress->id, [$path]);
    $job->handle($extractor);

    $waitress->refresh();
    expect($waitress->menu_image_paths)->toBe([$path])
        ->and($waitress->menuItems)->toHaveCount(0);
});

test('job updates waitress with menu items and image paths', function () {
    $user = User::factory()->create();
    $waitress = $user->waitresses()->create([
        'name' => 'Test',
        'slug' => 'test',
        'context' => '',
        'tools' => [['type' => 'send_email', 'name' => 'n', 'target' => 't@t.com']],
    ]);
    $path = "waitresses/{$waitress->id}/menu-extraction/menu.jpg";
    Storage::put($path, UploadedFile::fake()->image('menu.jpg', 10, 10)->get());

    $extractor = $this->mock(MenuExtractionService::class);
    $extractor->shouldReceive('extractMenuFromFiles')->once()->andReturn([
        'currency' => 'EUR',
        'items' => [
            ['name' => 'Pizza', 'category' => 'Mains', 'unit_price' => 12.00],
            ['name' => 'Water', 'category' => 'Drinks', 'unit_price' => 2.00],
        ],
    ]);

    $job = new ExtractMenuJob($waitress->id, [$path]);
    $job->handle($extractor);

    $waitress->refresh();
    expect($waitress->menu_image_paths)->toBe([$path])
        ->and($waitress->menu_currency)->toBe('EUR')
        ->and($waitress->menuItems)->toHaveCount(2);
});
