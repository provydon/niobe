import { queryParams, type RouteQueryOptions, type RouteDefinition, type RouteFormDefinition, applyUrlDefaults } from './../../../wayfinder'
/**
* @see \App\Http\Controllers\MenuItemController::store
* @see app/Http/Controllers/MenuItemController.php:12
* @route '/waitresses/{waitress}/menu-items'
*/
export const store = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'post'> => ({
    url: store.url(args, options),
    method: 'post',
})

store.definition = {
    methods: ["post"],
    url: '/waitresses/{waitress}/menu-items',
} satisfies RouteDefinition<["post"]>

/**
* @see \App\Http\Controllers\MenuItemController::store
* @see app/Http/Controllers/MenuItemController.php:12
* @route '/waitresses/{waitress}/menu-items'
*/
store.url = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions) => {
    if (typeof args === 'string' || typeof args === 'number') {
        args = { waitress: args }
    }

    if (typeof args === 'object' && !Array.isArray(args) && 'id' in args) {
        args = { waitress: args.id }
    }

    if (Array.isArray(args)) {
        args = {
            waitress: args[0],
        }
    }

    args = applyUrlDefaults(args)

    const parsedArgs = {
        waitress: typeof args.waitress === 'object'
        ? args.waitress.id
        : args.waitress,
    }

    return store.definition.url
            .replace('{waitress}', parsedArgs.waitress.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\MenuItemController::store
* @see app/Http/Controllers/MenuItemController.php:12
* @route '/waitresses/{waitress}/menu-items'
*/
store.post = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'post'> => ({
    url: store.url(args, options),
    method: 'post',
})

/**
* @see \App\Http\Controllers\MenuItemController::store
* @see app/Http/Controllers/MenuItemController.php:12
* @route '/waitresses/{waitress}/menu-items'
*/
const storeForm = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: store.url(args, options),
    method: 'post',
})

/**
* @see \App\Http\Controllers\MenuItemController::store
* @see app/Http/Controllers/MenuItemController.php:12
* @route '/waitresses/{waitress}/menu-items'
*/
storeForm.post = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: store.url(args, options),
    method: 'post',
})

store.form = storeForm

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
export const update = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteDefinition<'put'> => ({
    url: update.url(args, options),
    method: 'put',
})

update.definition = {
    methods: ["put","patch"],
    url: '/waitresses/{waitress}/menu-items/{menu_item}',
} satisfies RouteDefinition<["put","patch"]>

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
update.url = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions) => {
    if (Array.isArray(args)) {
        args = {
            waitress: args[0],
            menu_item: args[1],
        }
    }

    args = applyUrlDefaults(args)

    const parsedArgs = {
        waitress: typeof args.waitress === 'object'
        ? args.waitress.id
        : args.waitress,
        menu_item: typeof args.menu_item === 'object'
        ? args.menu_item.id
        : args.menu_item,
    }

    return update.definition.url
            .replace('{waitress}', parsedArgs.waitress.toString())
            .replace('{menu_item}', parsedArgs.menu_item.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
update.put = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteDefinition<'put'> => ({
    url: update.url(args, options),
    method: 'put',
})

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
update.patch = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteDefinition<'patch'> => ({
    url: update.url(args, options),
    method: 'patch',
})

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
const updateForm = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: update.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'PUT',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
updateForm.put = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: update.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'PUT',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

/**
* @see \App\Http\Controllers\MenuItemController::update
* @see app/Http/Controllers/MenuItemController.php:34
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
updateForm.patch = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: update.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'PATCH',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

update.form = updateForm

/**
* @see \App\Http\Controllers\MenuItemController::destroy
* @see app/Http/Controllers/MenuItemController.php:57
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
export const destroy = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteDefinition<'delete'> => ({
    url: destroy.url(args, options),
    method: 'delete',
})

destroy.definition = {
    methods: ["delete"],
    url: '/waitresses/{waitress}/menu-items/{menu_item}',
} satisfies RouteDefinition<["delete"]>

/**
* @see \App\Http\Controllers\MenuItemController::destroy
* @see app/Http/Controllers/MenuItemController.php:57
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
destroy.url = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions) => {
    if (Array.isArray(args)) {
        args = {
            waitress: args[0],
            menu_item: args[1],
        }
    }

    args = applyUrlDefaults(args)

    const parsedArgs = {
        waitress: typeof args.waitress === 'object'
        ? args.waitress.id
        : args.waitress,
        menu_item: typeof args.menu_item === 'object'
        ? args.menu_item.id
        : args.menu_item,
    }

    return destroy.definition.url
            .replace('{waitress}', parsedArgs.waitress.toString())
            .replace('{menu_item}', parsedArgs.menu_item.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\MenuItemController::destroy
* @see app/Http/Controllers/MenuItemController.php:57
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
destroy.delete = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteDefinition<'delete'> => ({
    url: destroy.url(args, options),
    method: 'delete',
})

/**
* @see \App\Http\Controllers\MenuItemController::destroy
* @see app/Http/Controllers/MenuItemController.php:57
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
const destroyForm = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: destroy.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'DELETE',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

/**
* @see \App\Http\Controllers\MenuItemController::destroy
* @see app/Http/Controllers/MenuItemController.php:57
* @route '/waitresses/{waitress}/menu-items/{menu_item}'
*/
destroyForm.delete = (args: { waitress: number | { id: number }, menu_item: number | { id: number } } | [waitress: number | { id: number }, menu_item: number | { id: number } ], options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: destroy.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'DELETE',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

destroy.form = destroyForm

const menuItems = {
    store: Object.assign(store, store),
    update: Object.assign(update, update),
    destroy: Object.assign(destroy, destroy),
}

export default menuItems