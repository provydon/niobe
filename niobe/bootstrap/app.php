<?php

use App\Http\Middleware\ForceHttpsMiddleware;
use App\Http\Middleware\HandleAppearance;
use App\Http\Middleware\HandleInertiaRequests;
use Illuminate\Auth\Access\AuthorizationException;
use Illuminate\Auth\AuthenticationException;
use Illuminate\Database\Eloquent\ModelNotFoundException;
use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;
use Illuminate\Http\Middleware\AddLinkHeadersForPreloadedAssets;
use Illuminate\Validation\ValidationException;
use Symfony\Component\HttpKernel\Exception\HttpException;
use Symfony\Component\HttpKernel\Exception\NotFoundHttpException;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        api: __DIR__.'/../routes/api.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware): void {
        $middleware->encryptCookies(except: ['appearance', 'sidebar_state']);

        $middleware->web(append: [
            HandleAppearance::class,
            HandleInertiaRequests::class,
            AddLinkHeadersForPreloadedAssets::class,
        ]);

        // Stateful API: same-origin SPA uses session cookies for /api/* (like boi-online-portal, fikets)
        $middleware->statefulApi();

        // When behind a load balancer / reverse proxy (e.g. production)
        $middleware->trustProxies(at: '*');

        $middleware->web(prepend: [ForceHttpsMiddleware::class]);
    })
    ->withExceptions(function (Exceptions $exceptions): void {
        $exceptions->render(function (Throwable $e, $request) {
            // API / JSON: return consistent JSON errors for stateless clients
            if ($request->is('api/*') || $request->expectsJson()) {
                $status = method_exists($e, 'getStatusCode') ? $e->getStatusCode() : 500;

                if ($e instanceof ValidationException) {
                    return response()->json(['message' => $e->getMessage(), 'errors' => $e->errors()], 422);
                }
                if ($e instanceof AuthenticationException) {
                    return response()->json(['message' => $e->getMessage() ?: 'Unauthenticated.'], 401);
                }
                if ($e instanceof AuthorizationException) {
                    return response()->json(['message' => $e->getMessage() ?: 'Unauthorized.'], 403);
                }
                if ($e instanceof NotFoundHttpException
                    || $e instanceof ModelNotFoundException
                ) {
                    return response()->json(['message' => 'Resource not found.'], 404);
                }
                if ($e instanceof HttpException) {
                    return response()->json([
                        'message' => $e->getMessage() ?: 'An error occurred.',
                    ], $status);
                }

                return response()->json([
                    'message' => config('app.debug') ? $e->getMessage() : 'An error occurred.',
                ], $status >= 100 && $status < 600 ? $status : 500);
            }

            // Inertia: redirect to login when unauthenticated on web
            if ($e instanceof AuthenticationException && $request->header('X-Inertia')) {
                return Inertia::location(route('login'));
            }
        });
    })->create();
