<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Order;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class OrderController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        $orders = Order::query()
            ->whereHas('waitress', fn ($q) => $q->where('user_id', $request->user()->id))
            ->with('waitress:id,name,slug')
            ->latest('sent_at')
            ->paginate(20);

        return response()->json($orders);
    }
}
