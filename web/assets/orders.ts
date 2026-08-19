import { Order } from './types/order';

export const ORDERS: Order[] = [
  {
    id: '1',
    item: 'Order 1',
    details: 'Details about order 1',
    status: 'Pending',
    slug: 'order-1',
    date: '2024-07-01',
    items: [
      {
        product: {
          id: 1,
          title: 'Product 1',
          slug: 'product-1',
          imagesUrl: [require('../assets/images/i-phone-1.jpg')],
          price: 10.0,
          heroImage: require('../assets/images/i-phone-1.jpg'),
          category: {
            imageUrl: require('../assets/images/i-phone-1.jpg'),
            name: 'Phones',
            slug: 'phones',
          },
          maxQuantity: 10,
        },
        quantity: 1,
      },
      {
        product: {
          id: 2,
          title: 'Product 2',
          slug: 'product-2',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 20.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 2,
      },
    ],
  },
  {
    id: '2',
    item: 'Order 2',
    details: 'Details about order 2',
    status: 'Completed',
    slug: 'order-2',
    date: '2024-07-02',
    items: [
      {
        product: {
          id: 3,
          title: 'Product 3',
          slug: 'product-3',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 30.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 1,
      },
      {
        product: {
          id: 4,
          title: 'Product 4',
          slug: 'product-4',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 40.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 3,
      },
    ],
  },
  {
    id: '3',
    item: 'Order 3',
    details: 'Details about order 3',
    status: 'Shipped',
    slug: 'order-3',
    date: '2024-07-03',
    items: [
      {
        product: {
          id: 5,
          title: 'Product 5',
          slug: 'product-5',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 50.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 1,
      },
      {
        product: {
          id: 6,
          title: 'Product 6',
          slug: 'product-6',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 60.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 2,
      },
    ],
  },
  {
    id: '4',
    item: 'Order 4',
    details: 'Details about order 4',
    status: 'InTransit',
    slug: 'order-4',
    date: '2024-07-04',
    items: [
      {
        product: {
          id: 7,
          title: 'Product 7',
          slug: 'product-7',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 70.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 1,
      },
      {
        product: {
          id: 8,
          title: 'Product 8',
          slug: 'product-8',
          imagesUrl: [require('../assets/images/head-set-1.jpg')],
          price: 80.0,
          heroImage: require('../assets/images/head-set-1.jpg'),
          category: {
            imageUrl: require('../assets/images/head-set-1.jpg'),
            name: 'Accessories',
            slug: 'accessories',
          },
          maxQuantity: 10,
        },
        quantity: 1,
      },
    ],
  },
];
