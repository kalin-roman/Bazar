import { registerRootComponent } from 'expo';

import App from './src/app/(shop)';

// registerRootComponent calls AppRegistry.registerComponent('main', () => App);
// It also ensures that whether you load the rapp in Expo Go or in a native build,
// the environment is set up appropriately
registerRootComponent(App);
