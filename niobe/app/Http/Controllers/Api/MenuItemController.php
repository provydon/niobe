<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\MenuItem;
use App\Models\Waitress;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class MenuItemController extends Controller
{
    public function index(Waitress $waitress): JsonResponse
    {
        $this->authorize('update', $waitress);
        $items = $waitress->menuItems()->orderBy('position')->get();

        return response()->json($items);
    }

    public function store(Request $request, Waitress $waitress): JsonResponse
    {
        $this->authorize('update', $waitress);

        $validated = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'category' => ['nullable', 'string', 'max:100'],
            'unit_price' => ['nullable', 'numeric', 'min:0'],
        ]);

        $position = $waitress->menuItems()->max('position') + 1;

        $item = $waitress->menuItems()->create([
            'name' => trim($validated['name']),
            'category' => trim($validated['category'] ?? 'Other'),
            'unit_price' => (float) ($validated['unit_price'] ?? 0),
            'position' => $position,
        ]);

        return response()->json($item, 201);
    }

    public function update(Request $request, Waitress $waitress, MenuItem $menu_item): JsonResponse
    {
        $this->authorize('update', $waitress);

        if ($menu_item->waitress_id !== $waitress->id) {
            abort(404);
        }

        $validated = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'category' => ['nullable', 'string', 'max:100'],
            'unit_price' => ['nullable', 'numeric', 'min:0'],
        ]);

        $menu_item->update([
            'name' => trim($validated['name']),
            'category' => trim($validated['category'] ?? 'Other'),
            'unit_price' => (float) ($validated['unit_price'] ?? 0),
        ]);

        return response()->json($menu_item);
    }

    public function destroy(Waitress $waitress, MenuItem $menu_item): JsonResponse
    {
        $this->authorize('update', $waitress);

        if ($menu_item->waitress_id !== $waitress->id) {
            abort(404);
        }

        $menu_item->delete();

        return response()->json(['message' => __('Menu item removed.')]);
    }
}
