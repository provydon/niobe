<?php

namespace App\Policies;

use App\Models\User;
use App\Models\Waitress;
use Illuminate\Auth\Access\Response;

class WaitressPolicy
{
    public function viewAny(User $user): bool
    {
        return true;
    }

    public function view(User $user, Waitress $waitress): bool
    {
        return $waitress->user_id === $user->id;
    }

    public function create(User $user): bool
    {
        return true;
    }

    public function update(User $user, Waitress $waitress): bool
    {
        return $waitress->user_id === $user->id;
    }

    public function delete(User $user, Waitress $waitress): bool
    {
        return $waitress->user_id === $user->id;
    }

    public function restore(User $user, Waitress $waitress): bool
    {
        return false;
    }

    public function forceDelete(User $user, Waitress $waitress): bool
    {
        return false;
    }
}
