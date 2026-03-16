<?php

namespace App\Http\Controllers;

use App\Models\MenuItem;
use App\Models\Waitress;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;

class MenuItemController extends Controller
{
    /**
     * List menu items (JSON) for polling when menu is extracting.
     */
    public function index(Waitress $waitress): JsonResponse
    {
        $this->authorize('update', $waitress);

        $items = $waitress->menuItems()->orderBy('position')->get();

        return response()->json($items);
    }

    public function store(Request $request, Waitress $waitress): RedirectResponse
    {
        $this->authorize('update', $waitress);

        $validated = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'category' => ['nullable', 'string', 'max:100'],
            'unit_price' => ['nullable', 'numeric', 'min:0'],
        ]);

        $position = $waitress->menuItems()->max('position') + 1;

        $waitress->menuItems()->create([
            'name' => trim($validated['name']),
            'category' => trim($validated['category'] ?? 'Other'),
            'unit_price' => (float) ($validated['unit_price'] ?? 0),
            'position' => $position,
        ]);

        return redirect()->route('waitresses.edit', $waitress)->with('success', __('Menu item added.'));
    }

    public function update(Request $request, Waitress $waitress, MenuItem $menu_item): RedirectResponse
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

        return redirect()->route('waitresses.edit', $waitress)->with('success', __('Menu item updated.'));
    }

    public function destroy(Waitress $waitress, MenuItem $menu_item): RedirectResponse
    {
        $this->authorize('update', $waitress);

        if ($menu_item->waitress_id !== $waitress->id) {
            abort(404);
        }

        $menu_item->delete();

        return redirect()->route('waitresses.edit', $waitress)->with('success', __('Menu item removed.'));
    }
}
