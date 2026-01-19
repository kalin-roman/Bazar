import { Stack } from "expo-router";
import { Toast, ToastProvider } from "react-native-toast-notifications";
import Auth from "./auth";
import AuthProvider from "../provider/auth-provider";

export default function RootLayout() {
  return (
    <ToastProvider>
      <AuthProvider>
        <Stack>
          <Stack.Screen
            name="(shop)"
            options={{ headerShown: false, headerTitle: "Shop" }}
          />
          <Stack.Screen
            name="product"
            options={{ headerShown: false, headerTitle: "Product" }}
          />

          <Stack.Screen
            name="categories"
            options={{ headerShown: false, headerTitle: "Categories" }}
          />
          <Stack.Screen
            name="cart"
            options={{ presentation: "modal", headerTitle: "Shopping Cart" }}
          />
          <Stack.Screen name="auth" options={{ headerShown: false }} />
        </Stack>
      </AuthProvider>
    </ToastProvider>
  );
}

// export default RootLayout;
