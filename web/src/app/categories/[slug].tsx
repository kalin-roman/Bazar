import { ActivityIndicator, FlatList, Text, Image, View, StyleSheet } from "react-native";
import { Redirect, Stack, useLocalSearchParams } from "expo-router";
import { useEffect } from "react";
import { useCatalogStore } from "../../store/catalog-store";
import ProductListItem from "../../components/product-list-item";

interface CategoryProps {}

const Category = (props: CategoryProps) => {
  const { slug } = useLocalSearchParams<{ slug: string }>();
  const { categories, products, loading, fetch } = useCatalogStore();

  useEffect(() => {
    if (categories.length === 0) fetch();
  }, []);

  if (loading && categories.length === 0) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  const category = categories.find((category) => category.Slug === slug);

  if (!category) return <Redirect href="/404" />;

  const categoryProducts = products.filter((product) => product.CategoryID === category.ID);

  return (
    <View style={styles.container}>
      <Stack.Screen options={{ title: category.Name }} />
      <Image source={{ uri: category.ImageURL }} style={styles.categoryImage} />
      <Text style={styles.categoryName}>{category.Name}</Text>
      <FlatList
        data={categoryProducts}
        keyExtractor={(item) => item.ID.toString()}
        renderItem={({ item }) => ( <ProductListItem product={item}/>)}
        numColumns={2}
        columnWrapperStyle={styles.productRow}
        contentContainerStyle={styles.productsList}
      />
    </View>
  );
};

export default Category;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    padding: 16,
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  categoryImage: {
    width: "100%",
    height: 200,
    resizeMode: "cover",
    borderRadius: 8,
    marginBottom: 16,
  },
  categoryName: {
    fontSize: 24,
    fontWeight: "bold",
    marginBottom: 16,
  },
  productsList: {
    flexGrow: 1,
  },
  productRow: {
    justifyContent: "space-between",
  },
});
