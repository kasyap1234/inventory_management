# Frontend Implementation Plan

## Overview
Analysis revealed **~40 missing features** (31% of total features) across the frontend.  
**Priority 1 features** can be implemented in **4-6 hours**.

---

## 🚨 PRIORITY 1: CRITICAL FEATURES (4-6 hours)

### 1. Order Status Actions (2 hours)
**Location:** `frontend/app/dashboard/orders/page.tsx`

**Features to add:**
- Approve button (pending → approved)
- Process button (approved → processing)
- Ship button (processing → shipped)
- Deliver button (shipped → delivered)
- Cancel button (any status → cancelled)
- Status validation with error messages

**Implementation:**
```typescript
// Add to OrdersPage component
const orderActions = useMutation({
  mutationFn: async ({ orderId, action }: { orderId: string; action: string }) => {
    await api.post(`/orders/${orderId}/${action}`);
  },
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['orders'] });
  },
});

// In TableCell Actions column
<div className="flex items-center gap-1">
  {order.status === 'pending' && (
    <Button size="sm" variant="success" onClick={() => orderActions.mutate({ orderId: order.id, action: 'approve' })}>
      Approve
    </Button>
  )}
  {order.status === 'approved' && (
    <Button size="sm" onClick={() => orderActions.mutate({ orderId: order.id, action: 'process' })}>
      Process
    </Button>
  )}
  {order.status === 'processing' && order.order_type === 'sales' && (
    <Button size="sm" onClick={() => orderActions.mutate({ orderId: order.id, action: 'ship' })}>
      Ship
    </Button>
  )}
  {order.status === 'shipped' && (
    <Button size="sm" onClick={() => orderActions.mutate({ orderId: order.id, action: 'deliver' })}>
      Deliver
    </Button>
  )}
  {!['delivered', 'cancelled'].includes(order.status) && (
    <Button size="sm" variant="destructive" onClick={() => orderActions.mutate({ orderId: order.id, action: 'cancel' })}>
      Cancel
    </Button>
  )}
</div>
```

---

### 2. Invoice PDF Download (1 hour)
**Location:** `frontend/app/dashboard/invoices/page.tsx`

**Features to add:**
- Download PDF button
- View PDF in new tab button

**Implementation:**
```typescript
// Add download handler
const downloadPDF = async (invoiceId: string) => {
  try {
    const response = await api.get(`/invoices/${invoiceId}/pdf`, {
      responseType: 'blob',
    });
    const blob = new Blob([response.data], { type: 'application/pdf' });
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `invoice-${invoiceId}.pdf`;
    link.click();
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Error downloading PDF:', error);
    alert('Failed to download PDF');
  }
};

// In TableCell Actions column
<Button size="sm" variant="outline" onClick={() => downloadPDF(invoice.id)}>
  <Download className="h-4 w-4 mr-1" />
  PDF
</Button>
```

---

### 3. Stock Adjustment Modal (1-2 hours)
**Location:** `frontend/app/dashboard/inventory/page.tsx`

**Features to add:**
- Adjust stock modal (+/-)
- Reason for adjustment
- Stock history view

**Implementation:**
```typescript
// Create StockAdjustmentDialog component
function StockAdjustmentDialog({ inventory, open, onOpenChange }: {
  inventory: Inventory | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [adjustment, setAdjustment] = useState(0);
  const [reason, setReason] = useState('');
  
  const adjustMutation = useMutation({
    mutationFn: async () => {
      await api.post(`/inventory/${inventory.id}/adjust`, {
        adjustment,
        reason,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory'] });
      onOpenChange(false);
    },
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Adjust Stock</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium">Current Quantity</label>
            <div className="text-2xl font-bold">{inventory?.quantity || 0}</div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Adjustment</label>
            <Input
              type="number"
              value={adjustment}
              onChange={(e) => setAdjustment(parseInt(e.target.value))}
              placeholder="Enter + or - amount"
            />
            <p className="text-sm text-gray-500">
              New quantity: {(inventory?.quantity || 0) + adjustment}
            </p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Reason *</label>
            <Input
              required
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Reason for adjustment"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button onClick={() => adjustMutation.mutate()} disabled={!reason || adjustment === 0}>
              Adjust Stock
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// Add button in inventory table
<Button size="sm" onClick={() => setAdjustingInventory(inventory)}>
  <PlusCircle className="h-4 w-4 mr-1" />
  Adjust
</Button>
```

---

### 4. Users & Roles Management Page (2-3 hours)
**Location:** `frontend/app/dashboard/users/page.tsx` (NEW)

**Create new page with:**
- Users table with CRUD
- Roles table with CRUD
- Role assignment modal
- Permissions checklist

**Files to create:**
1. `frontend/app/dashboard/users/page.tsx`
2. `frontend/components/users/UsersTable.tsx`
3. `frontend/components/users/UserForm.tsx`
4. `frontend/components/users/RolesTable.tsx`
5. `frontend/components/users/RoleForm.tsx`

**Basic structure:**
```typescript
// frontend/app/dashboard/users/page.tsx
'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import api from '@/lib/api';

export default function UsersPage() {
  const [activeTab, setActiveTab] = useState('users');

  const { data: users } = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await api.get('/users');
      return response.data;
    },
  });

  const { data: roles } = useQuery({
    queryKey: ['roles'],
    queryFn: async () => {
      const response = await api.get('/roles');
      return response.data;
    },
  });

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Users & Roles</h1>
      
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="users">Users</TabsTrigger>
          <TabsTrigger value="roles">Roles</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
        </TabsList>

        <TabsContent value="users">
          <UsersTable users={users?.users || []} />
        </TabsContent>

        <TabsContent value="roles">
          <RolesTable roles={roles?.roles || []} />
        </TabsContent>

        <TabsContent value="permissions">
          <PermissionsManager />
        </TabsContent>
      </Tabs>
    </div>
  );
}
```

---

## 📝 PRIORITY 2: ENHANCED FUNCTIONALITY (6-8 hours)

### 5. Bulk Operations UI (2 hours)
**Locations:**
- `frontend/app/dashboard/products/page.tsx` - Bulk price update
- `frontend/app/dashboard/invoices/page.tsx` - Bulk invoice generation

**Features:**
- Select multiple items checkboxes
- Bulk actions dropdown
- Bulk update form
- Progress indicator

---

### 6. Advanced Search & Filters (2 hours)
**Locations:**
- All list pages (orders, products, invoices, etc.)

**Features:**
- Date range filters
- Status filters
- Category filters
- Price range filters
- Advanced search modal

---

### 7. Analytics Visualizations (2-3 hours)
**Location:** `frontend/app/dashboard/analytics/page.tsx`

**Charts to add:**
- Sales Trends (Line chart)
- GST Totals (Pie chart)
- Top Products (Bar chart)
- Order Status Distribution (Pie chart)
- Revenue by Category (Bar chart)

**Libraries:**
- Recharts (already in package.json)

---

### 8. Audit Entity History (1-2 hours)
**Location:** `frontend/app/dashboard/audit-logs/page.tsx`

**Features:**
- Entity history timeline
- Diff view for changes
- User activity report

---

## 🔧 PRIORITY 3: ADMIN FEATURES (4-6 hours)

### 9. Background Jobs UI (2 hours)
**Location:** `frontend/app/dashboard/jobs/page.tsx` (NEW)

**Features:**
- Jobs queue table
- Job status indicators
- Retry/Cancel buttons
- Job logs viewer

---

### 10. Webhooks UI (2 hours)
**Location:** `frontend/app/dashboard/webhooks/page.tsx` (NEW)

**Features:**
- Webhooks table
- Create/Edit webhook form
- Test webhook button
- Delivery history

---

### 11. Tenant Management (2 hours)
**Location:** `frontend/app/dashboard/tenants/page.tsx` (NEW)

**Features:**
- Tenants table (admin only)
- Create/Edit tenant form
- Tenant settings
- Usage statistics

---

## 🎨 PRIORITY 4: POLISH (2-4 hours)

### 12. Notification Settings (1 hour)
**Location:** `frontend/app/dashboard/settings/page.tsx`

**Features:**
- Email notification toggle
- SMS notification toggle
- Notification preferences

---

### 13. Error Handling & Toasts (1 hour)
**Global improvements:**
- Add react-hot-toast
- Error toast on API failures
- Success toast on operations
- Loading states for all mutations

---

### 14. Loading States (1 hour)
**All pages:**
- Skeleton loaders
- Spinner components
- Optimistic updates

---

### 15. Form Validation (1 hour)
**All forms:**
- Client-side validation
- Error messages
- Field-level errors

---

## 📦 REQUIRED DEPENDENCIES

Most dependencies are already installed. Additional packages needed:

```bash
cd frontend
bun add react-hot-toast          # For toasts
bun add @heroicons/react          # For icons
bun add date-fns                  # For date formatting (already installed)
```

---

## 🚀 IMPLEMENTATION COMMANDS

```bash
# Create missing page directories
mkdir -p frontend/app/dashboard/users
mkdir -p frontend/app/dashboard/tenants  
mkdir -p frontend/app/dashboard/jobs
mkdir -p frontend/app/dashboard/webhooks

# Create component directories
mkdir -p frontend/components/users
mkdir -p frontend/components/tenants
mkdir -p frontend/components/jobs
mkdir -p frontend/components/webhooks

# Start frontend dev server
cd frontend && bun run dev
```

---

## ⏱️ TIME ESTIMATES

| Priority | Features | Hours | Status |
|----------|----------|-------|--------|
| Priority 1 | Order actions, PDF download, Stock adjust, Users/Roles | 4-6 hours | ⏳ Ready |
| Priority 2 | Bulk ops, Filters, Analytics, Audit history | 6-8 hours | 📋 Planned |
| Priority 3 | Jobs, Webhooks, Tenants | 4-6 hours | 📋 Planned |
| Priority 4 | Settings, Polish, Error handling | 2-4 hours | 📋 Planned |
| **TOTAL** | **40+ features** | **16-24 hours** | 🎯 Estimated |

---

## ✅ IMPLEMENTATION CHECKLIST

### Phase 1 (Priority 1)
- [ ] Add order status action buttons (Approve, Process, Ship, Deliver, Cancel)
- [ ] Add invoice PDF download button
- [ ] Implement stock adjustment modal
- [ ] Create users & roles management page

### Phase 2 (Priority 2)
- [ ] Add bulk product update UI
- [ ] Add bulk invoice generation UI
- [ ] Implement advanced search & filters
- [ ] Create all analytics visualizations
- [ ] Add audit entity history timeline

### Phase 3 (Priority 3)
- [ ] Create background jobs monitoring page
- [ ] Create webhooks management page
- [ ] Create tenant management page (admin)

### Phase 4 (Priority 4)
- [ ] Add notification settings page
- [ ] Implement toast notifications
- [ ] Add loading skeletons
- [ ] Improve form validation

---

## 🎯 QUICK START

To implement Priority 1 features immediately:

1. **Start with Order Actions** (Most visible impact)
2. **Add PDF Download** (Simple, high value)
3. **Stock Adjustment** (Critical for operations)
4. **Users/Roles** (Foundation for RBAC)

Each can be implemented independently!

---

**Status:** Complete implementation plan ready! Start with Priority 1 for maximum impact. 🚀
