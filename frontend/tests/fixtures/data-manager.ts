import { APIRequestContext } from '@playwright/test';
import {
  apiAdjustInventory,
  apiCreateDistributor,
  apiCreateInventory,
  apiCreateOrder,
  apiCreateProduct,
  apiCreateWarehouse,
  apiDeleteDistributor,
  apiDeleteOrder,
  apiDeleteProduct,
  apiDeleteWarehouse,
  apiListInventories,
  apiLogin,
} from './api-helpers';
import {
  testDistributors,
  testProducts,
  testUsers,
  testWarehouses,
} from './test-data';

type SeededEntities = {
  productId?: string;
  warehouseId?: string;
  distributorId?: string;
  inventoryId?: string;
  orderId?: string;
};

export class TestDataManager {
  private token: string = '';

  constructor(private readonly request: APIRequestContext) {}

  static async init(request: APIRequestContext) {
    const manager = new TestDataManager(request);
    await manager.authenticate();
    return manager;
  }

  async authenticate(
    email: string = testUsers.admin.email,
    password: string = testUsers.admin.password,
  ) {
    this.token = await apiLogin(this.request, email, password);
  }

  private ensureToken() {
    if (!this.token) {
      throw new Error('TestDataManager not authenticated');
    }
  }

  private extractId(entity: any): string | undefined {
    if (!entity) return undefined;
    return (
      entity.id ||
      entity.ID ||
      entity.product_id ||
      entity.warehouse_id ||
      entity.distributor_id ||
      entity.data?.id ||
      entity.data?.ID
    );
  }

  async createWarehouse() {
    this.ensureToken();
    const payload = {
      name: testWarehouses.basic.name,
      address: testWarehouses.basic.address,
      capacity: testWarehouses.basic.capacity,
    };
    const response = await apiCreateWarehouse(this.request, this.token, payload);
    const warehouse = response.warehouse || response;
    return {
      id: this.extractId(warehouse),
      payload,
    };
  }

  async createDistributor() {
    this.ensureToken();
    const payload = {
      name: testDistributors.basic.name,
      contact_email: testDistributors.basic.contactEmail,
      contact_phone: testDistributors.basic.contactPhone,
      address: testDistributors.basic.address,
    };
    const response = await apiCreateDistributor(this.request, this.token, payload);
    const distributor = response.distributor || response;
    return {
      id: this.extractId(distributor),
      payload,
    };
  }

  async createProduct() {
    this.ensureToken();
    const payload = {
      name: `${testProducts.basic.name} ${Date.now()}`,
      quantity: 20,
      unit_price: testProducts.basic.price,
      description: testProducts.basic.description,
    };
    const response = await apiCreateProduct(this.request, this.token, payload);
    const product = response.product || response;
    return {
      id: this.extractId(product),
      payload,
    };
  }

  async createInventory(productId: string, warehouseId: string, quantity = 20) {
    this.ensureToken();
    const payload = {
      product_id: productId,
      warehouse_id: warehouseId,
      quantity,
    };
    const response = await apiCreateInventory(this.request, this.token, payload);
    const inventory = response.inventory || response;
    return {
      id: this.extractId(inventory),
      payload,
    };
  }

  async adjustInventory(productId: string, adjustment: number, reason: string) {
    this.ensureToken();
    return apiAdjustInventory(this.request, this.token, {
      product_id: productId,
      adjustment,
      reason,
    });
  }

  async createOrder(params: {
    productId: string;
    warehouseId: string;
    distributorId: string;
    quantity?: number;
    unitPrice?: number;
    orderType?: 'sales' | 'purchase';
  }) {
    this.ensureToken();
    const payload = {
      order_type: params.orderType || 'sales',
      product_id: params.productId,
      warehouse_id: params.warehouseId,
      distributor_id: params.distributorId,
      quantity: params.quantity ?? 2,
      unit_price: params.unitPrice ?? 200,
      notes: 'E2E order seed',
    };
    const response = await apiCreateOrder(this.request, this.token, payload);
    const order = response.order || response;
    return {
      id: this.extractId(order),
      payload,
    };
  }

  async getInventoryQuantity(productId: string) {
    this.ensureToken();
    const response = await apiListInventories(this.request, this.token, { limit: 50 });
    const inventories =
      response.inventories ||
      response.inventory ||
      response.data?.inventories ||
      response.data ||
      [];
    const match = inventories.find((inv: any) => {
      const p = inv.product || {};
      return p.id === productId || p.ID === productId || inv.product_id === productId;
    });
    const quantity =
      match?.quantity ??
      match?.Quantity ??
      match?.stock ??
      (typeof match === 'number' ? match : 0);
    return { quantity: Number(quantity) || 0 };
  }

  async cleanup(ids: SeededEntities) {
    this.ensureToken();
    const tasks: Array<Promise<unknown>> = [];

    if (ids.orderId) {
      tasks.push(
        apiDeleteOrder(this.request, this.token, ids.orderId).catch(() => undefined),
      );
    }
    if (ids.productId) {
      tasks.push(
        apiDeleteProduct(this.request, this.token, ids.productId).catch(() => undefined),
      );
    }
    if (ids.distributorId) {
      tasks.push(
        apiDeleteDistributor(this.request, this.token, ids.distributorId).catch(
          () => undefined,
        ),
      );
    }
    if (ids.warehouseId) {
      tasks.push(
        apiDeleteWarehouse(this.request, this.token, ids.warehouseId).catch(
          () => undefined,
        ),
      );
    }

    await Promise.all(tasks);
  }
}

