import { queryParams, type RouteQueryOptions, type RouteDefinition, type RouteFormDefinition, applyUrlDefaults } from './../../wayfinder'
import menuItems from './menu-items'
/**
* @see \App\Http\Controllers\WaitressController::extractContext
* @see app/Http/Controllers/WaitressController.php:85
* @route '/waitresses/extract-context'
*/
export const extractContext = (options?: RouteQueryOptions): RouteDefinition<'post'> => ({
    url: extractContext.url(options),
    method: 'post',
})

extractContext.definition = {
    methods: ["post"],
    url: '/waitresses/extract-context',
} satisfies RouteDefinition<["post"]>

/**
* @see \App\Http\Controllers\WaitressController::extractContext
* @see app/Http/Controllers/WaitressController.php:85
* @route '/waitresses/extract-context'
*/
extractContext.url = (options?: RouteQueryOptions) => {
    return extractContext.definition.url + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::extractContext
* @see app/Http/Controllers/WaitressController.php:85
* @route '/waitresses/extract-context'
*/
extractContext.post = (options?: RouteQueryOptions): RouteDefinition<'post'> => ({
    url: extractContext.url(options),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::extractContext
* @see app/Http/Controllers/WaitressController.php:85
* @route '/waitresses/extract-context'
*/
const extractContextForm = (options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: extractContext.url(options),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::extractContext
* @see app/Http/Controllers/WaitressController.php:85
* @route '/waitresses/extract-context'
*/
extractContextForm.post = (options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: extractContext.url(options),
    method: 'post',
})

extractContext.form = extractContextForm

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
export const index = (options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: index.url(options),
    method: 'get',
})

index.definition = {
    methods: ["get","head"],
    url: '/waitresses',
} satisfies RouteDefinition<["get","head"]>

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
index.url = (options?: RouteQueryOptions) => {
    return index.definition.url + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
index.get = (options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: index.url(options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
index.head = (options?: RouteQueryOptions): RouteDefinition<'head'> => ({
    url: index.url(options),
    method: 'head',
})

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
const indexForm = (options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: index.url(options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
indexForm.get = (options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: index.url(options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::index
* @see app/Http/Controllers/WaitressController.php:62
* @route '/waitresses'
*/
indexForm.head = (options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: index.url({
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'HEAD',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'get',
})

index.form = indexForm

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
export const create = (options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: create.url(options),
    method: 'get',
})

create.definition = {
    methods: ["get","head"],
    url: '/waitresses/create',
} satisfies RouteDefinition<["get","head"]>

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
create.url = (options?: RouteQueryOptions) => {
    return create.definition.url + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
create.get = (options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: create.url(options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
create.head = (options?: RouteQueryOptions): RouteDefinition<'head'> => ({
    url: create.url(options),
    method: 'head',
})

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
const createForm = (options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: create.url(options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
createForm.get = (options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: create.url(options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::create
* @see app/Http/Controllers/WaitressController.php:75
* @route '/waitresses/create'
*/
createForm.head = (options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: create.url({
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'HEAD',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'get',
})

create.form = createForm

/**
* @see \App\Http\Controllers\WaitressController::store
* @see app/Http/Controllers/WaitressController.php:111
* @route '/waitresses'
*/
export const store = (options?: RouteQueryOptions): RouteDefinition<'post'> => ({
    url: store.url(options),
    method: 'post',
})

store.definition = {
    methods: ["post"],
    url: '/waitresses',
} satisfies RouteDefinition<["post"]>

/**
* @see \App\Http\Controllers\WaitressController::store
* @see app/Http/Controllers/WaitressController.php:111
* @route '/waitresses'
*/
store.url = (options?: RouteQueryOptions) => {
    return store.definition.url + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::store
* @see app/Http/Controllers/WaitressController.php:111
* @route '/waitresses'
*/
store.post = (options?: RouteQueryOptions): RouteDefinition<'post'> => ({
    url: store.url(options),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::store
* @see app/Http/Controllers/WaitressController.php:111
* @route '/waitresses'
*/
const storeForm = (options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: store.url(options),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::store
* @see app/Http/Controllers/WaitressController.php:111
* @route '/waitresses'
*/
storeForm.post = (options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: store.url(options),
    method: 'post',
})

store.form = storeForm

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
export const edit = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: edit.url(args, options),
    method: 'get',
})

edit.definition = {
    methods: ["get","head"],
    url: '/waitresses/{waitress}/edit',
} satisfies RouteDefinition<["get","head"]>

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
edit.url = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions) => {
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

    return edit.definition.url
            .replace('{waitress}', parsedArgs.waitress.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
edit.get = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: edit.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
edit.head = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'head'> => ({
    url: edit.url(args, options),
    method: 'head',
})

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
const editForm = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: edit.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
editForm.get = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: edit.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\WaitressController::edit
* @see app/Http/Controllers/WaitressController.php:159
* @route '/waitresses/{waitress}/edit'
*/
editForm.head = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: edit.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'HEAD',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'get',
})

edit.form = editForm

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
export const update = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'put'> => ({
    url: update.url(args, options),
    method: 'put',
})

update.definition = {
    methods: ["put","patch"],
    url: '/waitresses/{waitress}',
} satisfies RouteDefinition<["put","patch"]>

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
update.url = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions) => {
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

    return update.definition.url
            .replace('{waitress}', parsedArgs.waitress.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
update.put = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'put'> => ({
    url: update.url(args, options),
    method: 'put',
})

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
update.patch = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'patch'> => ({
    url: update.url(args, options),
    method: 'patch',
})

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
const updateForm = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: update.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'PUT',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
updateForm.put = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: update.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'PUT',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::update
* @see app/Http/Controllers/WaitressController.php:171
* @route '/waitresses/{waitress}'
*/
updateForm.patch = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
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
* @see \App\Http\Controllers\WaitressController::destroy
* @see app/Http/Controllers/WaitressController.php:198
* @route '/waitresses/{waitress}'
*/
export const destroy = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'delete'> => ({
    url: destroy.url(args, options),
    method: 'delete',
})

destroy.definition = {
    methods: ["delete"],
    url: '/waitresses/{waitress}',
} satisfies RouteDefinition<["delete"]>

/**
* @see \App\Http\Controllers\WaitressController::destroy
* @see app/Http/Controllers/WaitressController.php:198
* @route '/waitresses/{waitress}'
*/
destroy.url = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions) => {
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

    return destroy.definition.url
            .replace('{waitress}', parsedArgs.waitress.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\WaitressController::destroy
* @see app/Http/Controllers/WaitressController.php:198
* @route '/waitresses/{waitress}'
*/
destroy.delete = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteDefinition<'delete'> => ({
    url: destroy.url(args, options),
    method: 'delete',
})

/**
* @see \App\Http\Controllers\WaitressController::destroy
* @see app/Http/Controllers/WaitressController.php:198
* @route '/waitresses/{waitress}'
*/
const destroyForm = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: destroy.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'DELETE',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

/**
* @see \App\Http\Controllers\WaitressController::destroy
* @see app/Http/Controllers/WaitressController.php:198
* @route '/waitresses/{waitress}'
*/
destroyForm.delete = (args: { waitress: number | { id: number } } | [waitress: number | { id: number } ] | number | { id: number }, options?: RouteQueryOptions): RouteFormDefinition<'post'> => ({
    action: destroy.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'DELETE',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'post',
})

destroy.form = destroyForm

const waitresses = {
    extractContext: Object.assign(extractContext, extractContext),
    index: Object.assign(index, index),
    create: Object.assign(create, create),
    store: Object.assign(store, store),
    edit: Object.assign(edit, edit),
    update: Object.assign(update, update),
    destroy: Object.assign(destroy, destroy),
    menuItems: Object.assign(menuItems, menuItems),
}

export default waitresses