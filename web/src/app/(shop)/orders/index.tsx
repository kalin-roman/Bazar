import {
  ActivityIndicator,
  FlatList,
  Text,
  View,
  StyleSheet,
  ListRenderItem,
  Pressable,
} from "react-native";

import { useEffect } from "react";
import { Link, Stack } from "expo-router";
import { useOrdersStore } from "../../../store/orders-store";
import { ApiOrder } from "../../../lib/api";

// Backend-driven statuses aren't a fixed union the frontend controls
// anymore (only "Pending" exists in practice today — nothing updates
// an order's status yet — but the column is a plain text field, not
// an enum), so styling falls back to a neutral badge for anything
// unrecognized instead of assuming a closed set.
const statusColors: Record<string, string> = {
  Pending: "#ffcc00",
  Completed: "#4caf50",
  Shipped: "#2196f3",
  InTransit: "#ff9800",
};

interface OrdersProps {}

const Orders = (props: OrdersProps) => {
  const { orders, loading, error, fetch } = useOrdersStore();

  useEffect(() => {
    fetch();
  }, []);

  const renderItem: ListRenderItem<ApiOrder> = ({ item }) => {
    const totalCents = item.Items.reduce((sum, i) => sum + i.PriceCents * i.Quantity, 0);
    return (
      <Link href={`/orders/${item.ID}`} asChild>
        <Pressable style={styles.orderContainer}>
          <View style={styles.orderContent}>
            <View style={styles.orderDetailsContainer}>
              <Text style={styles.orderItem}>Order #{item.ID}</Text>
              <Text style={styles.orderDetails}>
                {item.Items.length} item(s) — ${(totalCents / 100).toFixed(2)}
              </Text>
              <Text style={styles.orderDate}>
                {new Date(item.CreatedAt).toLocaleDateString()}
              </Text>
            </View>
            <View
              style={[
                styles.statusBadge,
                { backgroundColor: statusColors[item.Status] ?? "#999" },
              ]}
            >
              <Text style={styles.statusText}>{item.Status}</Text>
            </View>
          </View>
        </Pressable>
      </Link>
    );
  };

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: "Orders" }} />
      {loading && orders.length === 0 ? (
        <View style={styles.centered}>
          <ActivityIndicator size="large" />
        </View>
      ) : error ? (
        <View style={styles.centered}>
          <Text style={styles.errorText}>Couldn't load orders: {error}</Text>
        </View>
      ) : orders.length === 0 ? (
        <View style={styles.centered}>
          <Text>No orders yet.</Text>
        </View>
      ) : (
        <FlatList
          data={orders}
          keyExtractor={(item) => item.ID.toString()}
          renderItem={renderItem}
        />
      )}
    </View>
  );
};

export default Orders;

const styles: { [key: string]: any } = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
  centered: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  errorText: {
    color: "#c00",
    textAlign: "center",
  },
  orderContainer: {
    backgroundColor: "#f8f8f8",
    padding: 16,
    marginVertical: 8,
    borderRadius: 8,
  },
  orderContent: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
  },
  orderDetailsContainer: {
    flex: 1,
  },
  orderItem: {
    fontSize: 18,
    fontWeight: "bold",
  },
  orderDetails: {
    fontSize: 14,
    color: "#555",
  },
  orderDate: {
    fontSize: 12,
    color: "#888",
    marginTop: 4,
  },
  statusBadge: {
    paddingVertical: 4,
    paddingHorizontal: 8,
    borderRadius: 4,
    alignSelf: "flex-start",
  },
  statusText: {
    fontSize: 12,
    fontWeight: "bold",
    color: "#fff",
  },
});
