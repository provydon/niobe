<?php

namespace App\Http\Controllers;

use App\Models\Order;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class DashboardController extends Controller
{
    public function __invoke(Request $request): Response
    {
        $totalOrders = Order::query()
            ->whereHas('waitress', fn ($q) => $q->where('user_id', $request->user()->id))
            ->count();

        return Inertia::render('Dashboard', [
            'totalOrders' => $totalOrders,
        ]);
    }
}
