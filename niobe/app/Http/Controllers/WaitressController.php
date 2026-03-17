<?php

namespace App\Http\Controllers;

use App\Models\Waitress;
use Inertia\Inertia;
use Inertia\Response;

class WaitressController extends Controller
{
    public function index(): Response
    {
        return Inertia::render('Waitresses/Index');
    }

    public function create(): Response
    {
        return Inertia::render('Waitresses/Create');
    }

    public function edit(Waitress $waitress): Response
    {
        $this->authorize('update', $waitress);

        return Inertia::render('Waitresses/Edit', [
            'waitressId' => $waitress->id,
        ]);
    }
}
