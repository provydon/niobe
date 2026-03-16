<?php

use App\Http\Controllers\MenuItemController;
use App\Http\Controllers\OAuthController;
use App\Http\Controllers\OrderController;
use App\Http\Controllers\PublicNiobeController;
use App\Http\Controllers\WaitressController;
use Illuminate\Support\Facades\Route;
use Laravel\Fortify\Features;

Route::prefix('oauth')->group(function () {
    Route::get('/{provider}/redirect', [OAuthController::class, 'redirect'])->middleware('guest')->name('oauth.redirect');
    Route::get('/{provider}/callback', [OAuthController::class, 'callback'])->name('oauth.callback');
});

Route::inertia('/', 'Welcome', [
    'canRegister' => Features::enabled(Features::registration()),
])->name('home');

Route::get('/n/{slug}', [PublicNiobeController::class, 'show'])->name('niobe.show');
Route::get('/n/{slug}/talk', [PublicNiobeController::class, 'talk'])->name('niobe.talk');
Route::get('/n/{slug}/menu-image/{index}', [PublicNiobeController::class, 'menuImage'])->name('niobe.menu-image')->where('index', '[0-9]+');

Route::middleware(['auth', 'verified'])->group(function () {
    Route::inertia('dashboard', 'Dashboard')->name('dashboard');
    Route::get('orders', [OrderController::class, 'index'])->name('orders.index');
    Route::post('waitresses/extract-context', [WaitressController::class, 'extractContext'])->name('waitresses.extract-context');
    Route::resource('waitresses', WaitressController::class)->except(['show']);
    Route::resource('waitresses.menu-items', MenuItemController::class)->only(['store', 'update', 'destroy']);
});

require __DIR__.'/settings.php';
