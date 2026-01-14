import * as React from 'react';
import { Text, View, StyleSheet } from 'react-native';

interface ProductProps {}

const Product = (props: ProductProps) => {
  return (
    <View style={styles.container}>
      <Text>Product</Text>
    </View>
  );
};

export default Product;

const styles = StyleSheet.create({
  container: {}
});
