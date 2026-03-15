<?php

namespace App\Http\Controllers;

use App\Models\User;
use Exception;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;
use Laravel\Socialite\Facades\Socialite;

class OAuthController extends Controller
{
    public function redirect(Request $request, string $provider)
    {
        return Socialite::driver($provider)->redirect();
    }

    public function callback(Request $request, string $provider)
    {
        if ($request->has('denied') || ! $request->has('code')) {
            return redirect()->route('login');
        }

        try {
            $socialUser = Socialite::driver($provider)->user();
            $user = User::where('email', $socialUser->getEmail())->first();

            if ($user) {
                $user->update([$provider.'_id' => $socialUser->getId()]);
            } else {
                $user = $this->createUser($socialUser, $provider);
            }

            Auth::login($user);

            return redirect()->intended(route('dashboard'));
        } catch (Exception $e) {
            Log::error('OAuth callback failed', ['error' => $e->getMessage(), 'provider' => $provider]);

            return redirect()->route('login')->with('error', 'Login failed. Please try again.');
        }
    }

    private function createUser($socialUser, string $provider): User
    {
        return User::create([
            'name' => $socialUser->getName(),
            'email' => $socialUser->getEmail(),
            $provider.'_id' => $socialUser->getId(),
            'password' => bcrypt(Str::random(32)),
        ]);
    }
}
