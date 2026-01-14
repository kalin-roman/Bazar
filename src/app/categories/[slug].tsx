import * as React from 'react';
import { Text, View, StyleSheet } from 'react-native';
import {Redirect, useLocalSearchParams} from 'expo-router';
import { CATEGORIES } from '../../../assets/categories';

interface CategoryProps {}

const Category = (props: CategoryProps) => {
  const { slug } = useLocalSearchParams<{ slug: string }>();

  const category = CATEGORIES.find(category => category.slug === slug);

  if (!category) return <Redirect href="/404" />;

  const products = category.products;

  return (
    <View style={styles.container}>
      <Text>Category</Text>
    </View>
  );
};

export default Category;

const styles = StyleSheet.create({
  container: {}
});
