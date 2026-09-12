import { Redirect, Stack, useLocalSearchParams } from "expo-router";
import { ActivityIndicator, TouchableOpacity, FlatList, Text, Image, StyleSheet, View } from "react-native";
import { useEffect, useState } from "react";

import { useToast } from "react-native-toast-notifications";
import { useCatalogStore } from "../../store/catalog-store";
import { useCartStore } from "../../store/cart-store";

const ProductDetails = () => {
    const { slug } = useLocalSearchParams<{ slug: string }>();
    const { products, loading, fetch } = useCatalogStore();
    // Hooks must run unconditionally on every render — the original
    // version of this screen called useCartStore() only after an
    // early `if (!product) return ...`, which breaks React's Rules of
    // Hooks (a real, latent crash risk when navigating between two
    // products, one found and one not). All hooks now run up front;
    // only the JSX below is conditional.
    const { items, addItem, incrementItem, decrementItem } = useCartStore();
    const toast = useToast();

    useEffect(() => {
        if (products.length === 0) fetch();
    }, []);

    const product = products.find((product) => product.Slug === slug);
    const cartItem = product ? items.find((item) => item.id === product.ID) : undefined;

    const [quantity, setQuantity] = useState(1);

    useEffect(() => {
        if (cartItem) setQuantity(cartItem.quantity);
    }, [cartItem?.quantity]);

    if (loading && products.length === 0) {
        return (
            <View style={styles.centered}>
                <ActivityIndicator size="large" />
            </View>
        );
    }

    if (!product) return <Redirect href='/404' />;

    const increaseQuantity = () => {

        if (quantity < product.MaxQuantity) {
            setQuantity(quantity + 1);
            incrementItem(product.ID);
        } else {
            toast.show("Cannnot add more items, reached maximum stock limit.", {
              type: "warning",
              placement: "bottom",
              duration: 1500
            });

    };
  };

    const decreaseQuantity = () => {
        if (quantity > 1) {
            setQuantity(quantity - 1);
            decrementItem(product.ID);
        }
    };

    const addToCart = () => {
        addItem({
            id: product.ID,
            title: product.Title,
            image: product.HeroImageURL,
            price: product.PriceCents / 100,
            quantity,
            maxQuantity: product.MaxQuantity,
        });
        toast.show('Item added to cart!', {
          type: "success",
          placement: "top",
          duration: 1500
        });
    };

    const priceEach = product.PriceCents / 100;
    const totalPrice = (priceEach * quantity).toFixed(2);

    return (
        <View style={styles.container}>
          <Stack.Screen options={{
            title: product.Title
            }} />
          <Image source={{ uri: product.HeroImageURL }} style={styles.heroImage}/>

          <View style={{padding:16, flex: 1}}>
            <Text style={styles.title}>Title: {product.Title}</Text>
            <Text style={styles.slug}>Slug: {product.Slug}</Text>
              <View style={styles.priceContainer}>
                  <Text style={styles.price}>Price: ${priceEach.toFixed(2)}</Text>
                  <Text style={styles.price}>Total: ${totalPrice}</Text>
              </View>

            <FlatList
              data={product.ImagesURL ?? []}
              keyExtractor={(item, index) => index.toString()}
              renderItem={({ item }) => (

                <Image source={{ uri: item }} style={styles.image} />
              )}
              horizontal
              showsHorizontalScrollIndicator={false}
              style={styles.imagesContainer}
            />
            <View style={styles.buttonContainer}>
                <TouchableOpacity
                style={styles.quantityButton}
                onPress={decreaseQuantity}
                disabled={quantity <= 1}
                >
                <Text style={styles.quantityButtonText}>-</Text>
                </TouchableOpacity>

                <Text style={styles.quantity}>{quantity}</Text>

                <TouchableOpacity
                style={styles.quantityButton}
                onPress={increaseQuantity}
                disabled={quantity >= product.MaxQuantity}
                >
                <Text style={styles.quantityButtonText}>+</Text>
                </TouchableOpacity>

                <TouchableOpacity
                style={[styles.addToCartButton,
                  {opacity: quantity === 0 ? 0.5 : 1},
                ]}
                onPress={addToCart}
                disabled={quantity === 0}>
                <Text style={styles.addToCartText}>Add to Cart</Text>
                </TouchableOpacity>

            </View>
            </View>
        </View>
    );
};
export default ProductDetails;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  heroImage: {
    width: '100%',
    height: 250,
    resizeMode: 'cover',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginVertical: 8,
  },
  slug: {
    fontSize: 18,
    color: '#555',
    marginBottom: 16,
  },
  priceContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 16,
  },
  price: {
    fontWeight: 'bold',
    color: '#000',
  },

  imagesContainer: {
    marginBottom: 16,
  },
  image: {
    width: 100,
    height: 100,
    marginRight: 8,
    borderRadius: 8,
  },
  buttonContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
    paddingHorizontal: 16,
  },
  quantityButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#007bff',
    alignItems: 'center',
    justifyContent: 'center',
    marginHorizontal: 8,
  },
  quantityButtonText: {
    fontSize: 24,
    color: '#fff',
  },
  quantity: {
    fontSize: 18,
    fontWeight: 'bold',
    marginHorizontal: 16,
  },
  addToCartButton: {
    flex: 1,
    backgroundColor: '#28a745',
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
    marginHorizontal: 8,
  },
  addToCartText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  errorMessage: {
    fontSize: 18,
    color: '#f00',
    textAlign: 'center',
    marginTop: 20,
  },
});
