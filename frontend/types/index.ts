export interface User {
  id: string;
  tenant_id: string;
  email: string;
  first_name: string;
  last_name: string;
  status: string;
  role?: string;
  created_at: string;
  updated_at: string;
}

export interface Tenant {
  id: string;
  name: string;
  subdomain: string;
  license_number?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  tenant_id: string;
  category_id?: string;
  name: string;
  batch_number?: string;
  expiry_date?: string;
  quantity: number;
  unit_price: number;
  barcode?: string;
  unit_of_measure?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  parent_id?: string;
  level: number;
  path: string;
  created_at: string;
  updated_at: string;
}

export interface Warehouse {
  id: string;
  tenant_id: string;
  name: string;
  address?: string;
  capacity?: number;
  license_number?: string;
  created_at: string;
  updated_at: string;
}

export interface Supplier {
  id: string;
  tenant_id: string;
  name: string;
  contact_email?: string;
  contact_phone?: string;
  address?: string;
  license_number?: string;
  created_at: string;
  updated_at: string;
}

export interface Distributor {
  id: string;
  tenant_id: string;
  name: string;
  contact_email?: string;
  contact_phone?: string;
  address?: string;
  license_number?: string;
  created_at: string;
  updated_at: string;
}

export interface Inventory {
  id: string;
  tenant_id: string;
  warehouse_id: string;
  product_id: string;
  quantity: number;
  last_updated: string;
}

export interface Order {
  id: string;
  tenant_id: string;
  order_type: 'purchase' | 'sales';
  supplier_id?: string;
  distributor_id?: string;
  product_id: string;
  warehouse_id: string;
  quantity: number;
  unit_price: number;
  status: 'pending' | 'approved' | 'received' | 'shipped' | 'delivered' | 'cancelled';
  order_date: string;
  expected_delivery?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Invoice {
  id: string;
  tenant_id: string;
  order_id: string;
  gstin?: string;
  hsn_sac?: string;
  taxable_amount?: number;
  gst_rate?: number;
  cgst?: number;
  sgst?: number;
  igst?: number;
  total_amount: number;
  status: 'unpaid' | 'paid' | 'overdue' | 'cancelled';
  issued_date: string;
  paid_date?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
  mfa_required?: boolean;
  mfa_token?: string;
  requires_2fa?: boolean;
  temp_token?: string;
  message?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface SignupData {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  tenant_name: string;
  subdomain: string;
}

export interface SignupResponse {
  user: User;
  message: string;
  verification_required: boolean;
}

export interface ForgotPasswordPayload {
  email: string;
}

export interface ResetPasswordPayload {
  token: string;
  password: string;
  confirm_password: string;
}

export interface Role {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  is_active?: boolean;
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  name: string;
  description?: string;
  resource: string;
  action: string;
  created_at: string;
}

export interface UserRole {
  user_id: string;
  role_id: string;
  assigned_at: string;
}

export interface RolePermission {
  role_id: string;
  permission_id: string;
  granted_at: string;
}
