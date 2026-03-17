<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class AuthController extends Controller
{
    /**
     * Issue a Sanctum token for the authenticated user (session auth).
     * Used by the SPA after Fortify login to obtain a token for stateless API calls.
     */
    public function token(Request $request): JsonResponse
    {
        $user = $request->user();
        $user->tokens()->where('name', 'spa')->delete();

        $token = $user->createToken('spa')->plainTextToken;

        return response()->json(['token' => $token]);
    }
}
