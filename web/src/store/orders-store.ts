import { create } from "zustand";
import { api, ApiOrder } from "../lib/api";

type OrdersState = {
  orders: ApiOrder[];
  loading: boolean;
  error: string | null;
  fetch: () => Promise<void>;
  checkout: (items: { product_id: number; quantity: number }[]) => Promise<ApiOrder>;
};

// Replaces the static assets/orders.ts INITIAL_ORDERS + local-only
// addOrder with real GET /orders (order history) and POST /orders
// (checkout) calls against the live backend.
export const useOrdersStore = create<OrdersState>((set, get) => ({
  orders: [],
  loading: false,
  error: null,
  fetch: async () => {
    set({ loading: true, error: null });
    try {
      const orders = await api.listOrders();
      set({ orders, loading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : "Failed to load orders",
        loading: false,
      });
    }
  },
  checkout: async (items) => {
    const created = await api.createOrder(items);
    set({ orders: [created, ...get().orders] });
    return created;
  },
}));
