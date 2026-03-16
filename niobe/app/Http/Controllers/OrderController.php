<?php

namespace App\Http\Controllers;

use App\Models\Order;
use Illuminate\Http\Request;
use Inertia\Inertia;
use Inertia\Response;

class OrderController extends Controller
{
    /**
     * List orders for waitresses owned by the authenticated user.
     */
    public function index(Request $request): Response
    {
        $orders = Order::query()
            ->whereHas('waitress', fn ($q) => $q->where('user_id', $request->user()->id))
            ->with('waitress:id,name,slug')
            ->latest('sent_at')
            ->paginate(20);

        return Inertia::render('Orders/Index', [
            'orders' => $orders,
        ]);
    }
}
