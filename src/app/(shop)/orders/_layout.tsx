import { Stack, Tabs } from "expo-router";
import { SafeAreaView } from "react-native-safe-area-context";

export default function OrdersLayout() {
  return (
    <Stack>
        {/* <Tabs.Screen name="shop" options={{ headerTitle: "Shop" }} /> */}
        {/* <Tabs.Screen name="auth" options={{ headerTitle: "Auth" }} /> */}
        {/* <Tabs.Screen name="cart" options={{ headerTitle: "Cart" }} /> */}
        <Stack.Screen name="index" options={{ headerShown: false }} />
    </Stack>
  );
};false

