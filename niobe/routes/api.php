<?php

use App\Http\Controllers\Api\DashboardController;
use App\Http\Controllers\Api\MenuItemController;
use App\Http\Controllers\Api\OrderController;
use App\Http\Controllers\Api\WaitressController;
use Illuminate\Support\Facades\Route;

Route::middleware(['auth:sanctum'])->group(function () {
    Route::get('/dashboard', DashboardController::class)->name('api.dashboard');

    Route::get('/orders', [OrderController::class, 'index'])->name('api.orders.index');

    Route::get('/waitresses/action-types', [WaitressController::class, 'actionTypes'])->name('api.waitresses.action-types');
    Route::post('/waitresses/extract-context', [WaitressController::class, 'extractContext'])->name('api.waitresses.extract-context');
    Route::get('/waitresses', [WaitressController::class, 'index'])->name('api.waitresses.index');
    Route::post('/waitresses', [WaitressController::class, 'store'])->name('api.waitresses.store');
    Route::get('/waitresses/{waitress}', [WaitressController::class, 'show'])->name('api.waitresses.show');
    Route::put('/waitresses/{waitress}', [WaitressController::class, 'update'])->name('api.waitresses.update');
    Route::delete('/waitresses/{waitress}', [WaitressController::class, 'destroy'])->name('api.waitresses.destroy');

    Route::get('/waitresses/{waitress}/menu-items', [MenuItemController::class, 'index'])->name('api.waitresses.menu-items.index');
    Route::post('/waitresses/{waitress}/menu-items', [MenuItemController::class, 'store'])->name('api.waitresses.menu-items.store');
    Route::put('/waitresses/{waitress}/menu-items/{menu_item}', [MenuItemController::class, 'update'])->name('api.waitresses.menu-items.update');
    Route::delete('/waitresses/{waitress}/menu-items/{menu_item}', [MenuItemController::class, 'destroy'])->name('api.waitresses.menu-items.destroy');
});
