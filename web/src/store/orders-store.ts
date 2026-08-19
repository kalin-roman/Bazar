import { create } from "zustand";
import { ORDERS as INITIAL_ORDERS } from "../../assets/orders";
import { Order } from "../../assets/types/order";

type OrdersState = {
  orders: Order[];
  addOrder: (order: Order) => void;
};

export const useOrdersStore = create<OrdersState>((set) => ({
  orders: INITIAL_ORDERS,
  addOrder: (order) =>
    set((state) => ({ orders: [order, ...state.orders] })),
}));
