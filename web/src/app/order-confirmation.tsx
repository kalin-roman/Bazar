import { Redirect, Stack, useLocalSearchParams, useRouter } from "expo-router";
import { FlatList, Image, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useOrdersStore } from "../store/orders-store";
import { useCatalogStore } from "../store/catalog-store";

const OrderConfirmation = () => {
  // Route param is named "slug" (unchanged file/route name), but its
  // value is the order's real numeric ID — backend orders have no
  // slug field.
  const { slug } = useLocalSearchParams<{ slug: string }>();
  const router = useRouter();
  const { orders } = useOrdersStore();
  const { products } = useCatalogStore();

  const order = orders.find((order) => order.ID.toString() === slug);

  if (!order) {
    return <Redirect href={"/404"} />;
  }

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: "Order Confirmed" }} />
      <Text style={styles.title}>Thank you for your order!</Text>
      <Text style={styles.subtitle}>Order #{order.ID}</Text>
      <Text style={styles.details}>
        {order.Items.length} item(s) — $
        {(order.Items.reduce((sum, i) => sum + i.PriceCents * i.Quantity, 0) / 100).toFixed(2)}
      </Text>
      <Text style={styles.date}>Placed on {new Date(order.CreatedAt).toLocaleDateString()}</Text>

      <Text style={styles.itemsTitle}>Items</Text>
      <FlatList
        data={order.Items}
        keyExtractor={(item) => item.ProductID.toString()}
        renderItem={({ item }) => {
          const product = products.find((p) => p.ID === item.ProductID);
          return (
            <View style={styles.orderItem}>
              {product && (
                <Image source={{ uri: product.HeroImageURL }} style={styles.heroImage} />
              )}
              <View style={styles.itemInfo}>
                <Text style={styles.itemName}>{product?.Title ?? `Product #${item.ProductID}`}</Text>
                <Text style={styles.itemPrice}>
                  ${(item.PriceCents / 100).toFixed(2)} x {item.Quantity}
                </Text>
              </View>
            </View>
          );
        }}
      />

      <View style={styles.footer}>
        <TouchableOpacity
          style={styles.button}
          onPress={() => router.replace("/orders")}
        >
          <Text style={styles.buttonText}>View Orders</Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.button, styles.secondaryButton]}
          onPress={() => router.replace("/")}
        >
          <Text style={styles.buttonText}>Continue Shopping</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
};

export default OrderConfirmation;

const styles: { [key: string]: any } = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: "#fff",
  },
  title: {
    fontSize: 24,
    fontWeight: "bold",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 18,
    fontWeight: "bold",
    marginBottom: 4,
  },
  details: {
    fontSize: 14,
    color: "#555",
  },
  date: {
    fontSize: 12,
    color: "#888",
    marginBottom: 16,
  },
  itemsTitle: {
    fontSize: 18,
    fontWeight: "bold",
    marginBottom: 8,
  },
  orderItem: {
    flexDirection: "row",
    alignItems: "center",
    marginTop: 8,
    padding: 16,
    backgroundColor: "#f8f8f8",
    borderRadius: 8,
  },
  heroImage: {
    width: 60,
    height: 60,
    borderRadius: 8,
    marginRight: 12,
  },
  itemInfo: {},
  itemName: {
    fontSize: 16,
    fontWeight: "bold",
  },
  itemPrice: {
    fontSize: 14,
    marginTop: 4,
    color: "#555",
  },
  footer: {
    marginTop: 24,
    gap: 12,
  },
  button: {
    backgroundColor: "#1BC464",
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: "center",
  },
  secondaryButton: {
    backgroundColor: "#333",
  },
  buttonText: {
    color: "#fff",
    fontSize: 16,
    fontWeight: "bold",
  },
});
