import OAuthController from './OAuthController'
import PublicNiobeController from './PublicNiobeController'
import WaitressController from './WaitressController'
import MenuItemController from './MenuItemController'
import Settings from './Settings'

const Controllers = {
    OAuthController: Object.assign(OAuthController, OAuthController),
    PublicNiobeController: Object.assign(PublicNiobeController, PublicNiobeController),
    WaitressController: Object.assign(WaitressController, WaitressController),
    MenuItemController: Object.assign(MenuItemController, MenuItemController),
    Settings: Object.assign(Settings, Settings),
}

export default Controllers