import { Stack } from "expo-router";

export default function RootLayout() {
  return (
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
    </Stack>
  );
}

// export default RootLayout;