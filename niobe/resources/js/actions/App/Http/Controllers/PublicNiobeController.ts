import { queryParams, type RouteQueryOptions, type RouteDefinition, type RouteFormDefinition, applyUrlDefaults } from './../../../../wayfinder'
/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
export const show = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: show.url(args, options),
    method: 'get',
})

show.definition = {
    methods: ["get","head"],
    url: '/n/{slug}',
} satisfies RouteDefinition<["get","head"]>

/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
show.url = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions) => {
    if (typeof args === 'string' || typeof args === 'number') {
        args = { slug: args }
    }

    if (Array.isArray(args)) {
        args = {
            slug: args[0],
        }
    }

    args = applyUrlDefaults(args)

    const parsedArgs = {
        slug: args.slug,
    }

    return show.definition.url
            .replace('{slug}', parsedArgs.slug.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
show.get = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: show.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
show.head = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteDefinition<'head'> => ({
    url: show.url(args, options),
    method: 'head',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
const showForm = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: show.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
showForm.get = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: show.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::show
* @see app/Http/Controllers/PublicNiobeController.php:16
* @route '/n/{slug}'
*/
showForm.head = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: show.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'HEAD',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'get',
})

show.form = showForm

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
export const talk = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: talk.url(args, options),
    method: 'get',
})

talk.definition = {
    methods: ["get","head"],
    url: '/n/{slug}/talk',
} satisfies RouteDefinition<["get","head"]>

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
talk.url = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions) => {
    if (typeof args === 'string' || typeof args === 'number') {
        args = { slug: args }
    }

    if (Array.isArray(args)) {
        args = {
            slug: args[0],
        }
    }

    args = applyUrlDefaults(args)

    const parsedArgs = {
        slug: args.slug,
    }

    return talk.definition.url
            .replace('{slug}', parsedArgs.slug.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
talk.get = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: talk.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
talk.head = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteDefinition<'head'> => ({
    url: talk.url(args, options),
    method: 'head',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
const talkForm = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: talk.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
talkForm.get = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: talk.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::talk
* @see app/Http/Controllers/PublicNiobeController.php:29
* @route '/n/{slug}/talk'
*/
talkForm.head = (args: { slug: string | number } | [slug: string | number ] | string | number, options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: talk.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'HEAD',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'get',
})

talk.form = talkForm

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
export const menuImage = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: menuImage.url(args, options),
    method: 'get',
})

menuImage.definition = {
    methods: ["get","head"],
    url: '/n/{slug}/menu-image/{index}',
} satisfies RouteDefinition<["get","head"]>

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
menuImage.url = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions) => {
    if (Array.isArray(args)) {
        args = {
            slug: args[0],
            index: args[1],
        }
    }

    args = applyUrlDefaults(args)

    const parsedArgs = {
        slug: args.slug,
        index: args.index,
    }

    return menuImage.definition.url
            .replace('{slug}', parsedArgs.slug.toString())
            .replace('{index}', parsedArgs.index.toString())
            .replace(/\/+$/, '') + queryParams(options)
}

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
menuImage.get = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions): RouteDefinition<'get'> => ({
    url: menuImage.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
menuImage.head = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions): RouteDefinition<'head'> => ({
    url: menuImage.url(args, options),
    method: 'head',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
const menuImageForm = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: menuImage.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
menuImageForm.get = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: menuImage.url(args, options),
    method: 'get',
})

/**
* @see \App\Http\Controllers\PublicNiobeController::menuImage
* @see app/Http/Controllers/PublicNiobeController.php:62
* @route '/n/{slug}/menu-image/{index}'
*/
menuImageForm.head = (args: { slug: string | number, index: string | number } | [slug: string | number, index: string | number ], options?: RouteQueryOptions): RouteFormDefinition<'get'> => ({
    action: menuImage.url(args, {
        [options?.mergeQuery ? 'mergeQuery' : 'query']: {
            _method: 'HEAD',
            ...(options?.query ?? options?.mergeQuery ?? {}),
        }
    }),
    method: 'get',
})

menuImage.form = menuImageForm

const PublicNiobeController = { show, talk, menuImage }

export default PublicNiobeController