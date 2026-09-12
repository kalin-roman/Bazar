import { Redirect, Stack, useLocalSearchParams } from "expo-router";
import { FlatList, StyleSheet, Text, View, Image, ActivityIndicator } from "react-native";
import { useEffect } from "react";
import { useOrdersStore } from "../../../store/orders-store";
import { useCatalogStore } from "../../../store/catalog-store";

const statusColors: Record<string, string> = {
  Pending: "orange",
  Completed: "green",
  Shipped: "blue",
  InTransit: "purple",
};

const OrderDetails = () => {
  // Route param is named "slug" (unchanged file/route name), but its
  // value is the order's real numeric ID.
  const { slug } = useLocalSearchParams<{ slug: string }>();

  const { orders, loading, fetch } = useOrdersStore();
  const { products, fetch: fetchCatalog } = useCatalogStore();

  useEffect(() => {
    if (orders.length === 0) fetch();
    if (products.length === 0) fetchCatalog();
  }, []);

  if (loading && orders.length === 0) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const order = orders.find((order) => order.ID.toString() === slug);

  if (!order) {
    return <Redirect href={"/404"} />;
  }

  const totalCents = order.Items.reduce((sum, i) => sum + i.PriceCents * i.Quantity, 0);

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: `Order #${order.ID}` }} />
      <Text style={styles.item}>Order #{order.ID}</Text>
      <Text style={styles.details}>
        {order.Items.length} item(s) — ${(totalCents / 100).toFixed(2)}
      </Text>
      <View style={[styles.statusBadge, { backgroundColor: statusColors[order.Status] ?? "#999" }]}>
        <Text style={styles.statusText}>{order.Status}</Text>
      </View>
      <Text style={styles.date}>Order Date: {new Date(order.CreatedAt).toLocaleDateString()}</Text>
      <Text style={styles.itemsTitle}>Items Order:</Text>
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
                <Text style={styles.itemSubtotal}>
                  ${((item.PriceCents * item.Quantity) / 100).toFixed(2)}
                </Text>
              </View>
            </View>
          );
        }}
      />
    </View>
  );
};
export default OrderDetails;

const styles: { [key: string]: any } = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#fff',
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  item: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  details: {
    fontSize: 16,
    marginBottom: 16,
  },
  statusBadge: {
    padding: 8,
    borderRadius: 4,
    alignSelf: 'flex-start',
  },
  statusText: {
    color: '#fff',
    fontWeight: 'bold',
  },
  date: {
    fontSize: 14,
    color: '#555',
    marginTop: 16,
  },
  itemsTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    marginTop: 16,
    marginBottom: 8,
  },
  orderItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 8,
    padding: 16,
    backgroundColor: '#f8f8f8',
    borderRadius: 8,
  },
  heroImage: {
    width: '50%',
    height: 100,
    borderRadius: 10,
  },
  itemInfo: {},
  itemName: {
    fontSize: 16,
    fontWeight: 'bold',
  },
  itemPrice: {
    fontSize: 14,
    marginTop: 4,
  },
  itemSubtotal: {
    fontSize: 14,
    fontWeight: 'bold',
    marginTop: 4,
  },
});
