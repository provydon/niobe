<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Order;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class DashboardController extends Controller
{
    public function __invoke(Request $request): JsonResponse
    {
        $totalOrders = Order::query()
            ->whereHas('waitress', fn ($q) => $q->where('user_id', $request->user()->id))
            ->count();

        return response()->json(['totalOrders' => $totalOrders]);
    }
}
