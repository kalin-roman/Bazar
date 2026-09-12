import { useEffect } from "react";
import { ActivityIndicator, FlatList, StyleSheet, Text, View } from "react-native";
import { useCatalogStore } from "../../store/catalog-store";
import ProductListItem from "../../components/product-list-item";
import { ListHeader } from "../../components/list-header";

const Home = () => {
  const { products, loading, error, fetch } = useCatalogStore();

  useEffect(() => {
    fetch();
  }, []);

  if (loading && products.length === 0) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.centered}>
        <Text style={styles.errorText}>Couldn't load products: {error}</Text>
      </View>
    );
  }

  return (
    <View>
      <FlatList
      data={products}
      renderItem={({ item }) => <ProductListItem product={item} />}
      keyExtractor={item => item.ID.toString()}
      numColumns={2}
      ListHeaderComponent={ListHeader}
      contentContainerStyle={styles.flatListContent}
      columnWrapperStyle={styles.flatListColumn}
      style={{ paddingHorizontal: 10, paddingVertical: 5 }}
      />
    </View>
  );
};

export default Home;

const styles = StyleSheet.create({
  flatListContent: {
    paddingBottom: 20,
  },
  flatListColumn: {
    justifyContent: 'space-between',
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 16,
  },
  errorText: {
    color: '#c00',
    textAlign: 'center',
  },
});
