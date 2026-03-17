<?php

namespace Tests;

use Illuminate\Foundation\Testing\TestCase as BaseTestCase;
use Laravel\Fortify\Features;

abstract class TestCase extends BaseTestCase
{
    /**
     * Assert redirect to the given route (compares path only so http vs https in tests do not matter).
     */
    protected function assertRedirectToRoute(mixed $response, string $name, array $parameters = []): void
    {
        $response->assertRedirect();
        $expectedPath = parse_url(route($name, $parameters), PHP_URL_PATH);
        $actualPath = parse_url($response->headers->get('Location'), PHP_URL_PATH);
        $this->assertSame($expectedPath, $actualPath, 'Redirect path does not match.');
    }

    protected function skipUnlessFortifyFeature(string $feature, ?string $message = null): void
    {
        if (! Features::enabled($feature)) {
            $this->markTestSkipped($message ?? "Fortify feature [{$feature}] is not enabled.");
        }
    }
}
