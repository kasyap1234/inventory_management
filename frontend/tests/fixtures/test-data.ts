/**
 * Test data fixtures for Playwright tests
 */

export const testUsers = {
  admin: {
    email: 'admin@test.com',
    password: 'Admin@123456',
    role: 'admin',
  },
  manager: {
    email: 'manager@test.com',
    password: 'Manager@123456',
    role: 'manager',
  },
  user: {
    email: 'user@test.com',
    password: 'User@123456',
    role: 'user',
  },
  newUser: {
    email: `test-${Date.now()}@example.com`,
    password: 'TestPassword@123',
    name: 'Test User',
    phoneNumber: '+1234567890',
  },
};

export const testProducts = {
  basic: {
    name: 'Test Product',
    sku: `TEST-${Date.now()}`,
    price: 100,
    category: 'Electronics',
    description: 'A test product for automated testing',
  },
  withInventory: {
    name: 'Test Product with Stock',
    sku: `TEST-INV-${Date.now()}`,
    price: 250,
    category: 'Electronics',
    description: 'Product with inventory',
    quantity: 100,
  },
};

export const testOrders = {
  simple: {
    items: [
      {
        productId: '',
        quantity: 5,
      },
    ],
    shippingAddress: '123 Test Street',
    notes: 'Test order',
  },
};

export const testCategories = {
  basic: {
    name: `Test Category ${Date.now()}`,
    description: 'Test category for automated testing',
  },
};

export const testWarehouses = {
  basic: {
    name: `Test Warehouse ${Date.now()}`,
    address: '123 Warehouse Street',
    capacity: 10000,
  },
};

export const apiEndpoints = {
  auth: {
    login: '/api/v1/auth/login',
    signup: '/api/v1/auth/signup',
    logout: '/api/v1/auth/logout',
    verifyEmail: '/api/v1/auth/verify-email',
    forgotPassword: '/api/v1/auth/forgot-password',
    resetPassword: '/api/v1/auth/reset-password',
  },
  products: {
    list: '/api/v1/products',
    create: '/api/v1/products',
    update: (id: string) => `/api/v1/products/${id}`,
    delete: (id: string) => `/api/v1/products/${id}`,
  },
  inventory: {
    list: '/api/v1/inventory',
    update: '/api/v1/inventory/adjust',
  },
  orders: {
    list: '/api/v1/orders',
    create: '/api/v1/orders',
  },
};

export const selectors = {
  auth: {
    emailInput: '#email',
    passwordInput: '#password',
    loginButton: 'button[type="submit"]',
    signupLink: 'a[href="/signup"]',
    forgotPasswordLink: 'a[href="/forgot-password"]',
    errorMessage: '[class*="bg-red"]',
    successMessage: '[class*="bg-green"]',
  },
  dashboard: {
    sidebar: '[role="navigation"]',
    productsLink: 'a[href*="/products"]',
    ordersLink: 'a[href*="/orders"]',
    inventoryLink: 'a[href*="/inventory"]',
    analyticsLink: 'a[href*="/analytics"]',
    profileMenu: '[data-testid="profile-menu"]',
    logoutButton: 'button:has-text("Logout")',
  },
  products: {
    createButton: 'button:has-text("Add Product")',
    searchInput: 'input[placeholder*="Search"]',
    productRow: '[data-testid^="product-row"]',
    editButton: 'button:has-text("Edit")',
    deleteButton: 'button:has-text("Delete")',
    confirmDelete: 'button:has-text("Confirm")',
  },
};
