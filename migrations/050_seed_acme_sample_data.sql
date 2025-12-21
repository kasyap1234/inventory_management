-- Seed realistic sample data for tenant acme (acme@gmail.com)
-- This script is idempotent and scoped to tenant ff5c2ea7-4d2d-4928-a34f-faac6be42c71
-- Run with:
-- PGPASSWORD=tespass psql -h localhost -p 5440 -U testuser -d testdb -f migrations/041_seed_acme_sample_data.sql

DO $$
DECLARE
    v_tenant         CONSTANT uuid := 'ff5c2ea7-4d2d-4928-a34f-faac6be42c71';
    v_user           CONSTANT uuid := '812ae9d8-6a2d-4382-a0ca-bacf044613d6';

    w_main           CONSTANT uuid := '11111111-1111-1111-1111-111111111111';
    w_east           CONSTANT uuid := '22222222-2222-2222-2222-222222222222';

    s_agri           CONSTANT uuid := '33333333-3333-3333-3333-333333333333';
    s_green          CONSTANT uuid := '33333333-3333-3333-3333-444444444444';

    d_farm           CONSTANT uuid := '44444444-4444-4444-4444-444444444444';
    d_coastal        CONSTANT uuid := '44444444-4444-4444-4444-555555555555';

    c_fertilizers    CONSTANT uuid := '55555555-5555-5555-5555-555555555555';
    c_irrigation     CONSTANT uuid := '55555555-5555-5555-5555-666666666666';
    c_seeds          CONSTANT uuid := '55555555-5555-5555-5555-777777777777';

    p_npk            CONSTANT uuid := '66666666-6666-6666-6666-666666666666';
    p_drip           CONSTANT uuid := '77777777-7777-7777-7777-777777777777';
    p_seed           CONSTANT uuid := '88888888-8888-8888-8888-888888888888';

    b_npk            CONSTANT uuid := '99999999-9999-9999-9999-111111111111';
    b_drip_a         CONSTANT uuid := '99999999-9999-9999-9999-222222222222';
    b_seed           CONSTANT uuid := '99999999-9999-9999-9999-333333333333';

    inv_npk_main     CONSTANT uuid := 'aaaaaaaa-aaaa-aaaa-aaaa-000000000001';
    inv_drip_main    CONSTANT uuid := 'aaaaaaaa-aaaa-aaaa-aaaa-000000000002';
    inv_seed_main    CONSTANT uuid := 'aaaaaaaa-aaaa-aaaa-aaaa-000000000003';
    inv_drip_east    CONSTANT uuid := 'aaaaaaaa-aaaa-aaaa-aaaa-000000000004';

    ord_po_1001      CONSTANT uuid := 'bbbbbbbb-bbbb-bbbb-bbbb-111111111111';
    ord_po_1002      CONSTANT uuid := 'bbbbbbbb-bbbb-bbbb-bbbb-222222222222';
    ord_so_2001      CONSTANT uuid := 'cccccccc-cccc-cccc-cccc-111111111111';
    ord_so_2002      CONSTANT uuid := 'cccccccc-cccc-cccc-cccc-222222222222';
    ord_so_2003      CONSTANT uuid := 'cccccccc-cccc-cccc-cccc-333333333333';

    oi_po_1001       CONSTANT uuid := 'dddddddd-dddd-dddd-dddd-111111111111';
    oi_po_1002       CONSTANT uuid := 'dddddddd-dddd-dddd-dddd-222222222222';
    oi_so_2001       CONSTANT uuid := 'dddddddd-dddd-dddd-dddd-333333333333';
    oi_so_2002       CONSTANT uuid := 'dddddddd-dddd-dddd-dddd-444444444444';
    oi_so_2003       CONSTANT uuid := 'dddddddd-dddd-dddd-dddd-555555555555';

    invc_po_1001     CONSTANT uuid := 'eeeeeeee-eeee-eeee-eeee-111111111111';
    invc_so_2003     CONSTANT uuid := 'eeeeeeee-eeee-eeee-eeee-222222222222';

    res_drip_hold    CONSTANT uuid := 'fafafafa-fafa-fafa-fafa-000000000001';
    adj_drip_damage  CONSTANT uuid := 'f1f1f1f1-f1f1-f1f1-f1f1-000000000001';

    notif_low_stock  CONSTANT uuid := 'abababab-abab-abab-abab-111111111111';
    notif_reorder    CONSTANT uuid := 'abababab-abab-abab-abab-222222222222';
    notif_shipment   CONSTANT uuid := 'abababab-abab-abab-abab-333333333333';

    tok_web          CONSTANT uuid := 'edededed-eded-eded-eded-111111111111';
BEGIN
    -- Safety check
    IF NOT EXISTS (SELECT 1 FROM tenants WHERE id = v_tenant) THEN
        RAISE EXCEPTION 'Tenant % not found. Seed aborted.', v_tenant;
    END IF;

    -- Clean existing Acme seed data
    DELETE FROM notification_deliveries WHERE tenant_id = v_tenant;
    DELETE FROM notifications WHERE tenant_id = v_tenant;
    DELETE FROM inventory_reservations WHERE tenant_id = v_tenant;
    DELETE FROM stock_adjustments WHERE tenant_id = v_tenant;
    DELETE FROM order_items WHERE tenant_id = v_tenant;
    DELETE FROM invoices WHERE tenant_id = v_tenant;
    DELETE FROM orders WHERE tenant_id = v_tenant;
    DELETE FROM inventory WHERE tenant_id = v_tenant;
    DELETE FROM batches WHERE tenant_id = v_tenant;
    DELETE FROM product_images WHERE tenant_id = v_tenant;
    DELETE FROM products WHERE tenant_id = v_tenant;
    DELETE FROM categories WHERE tenant_id = v_tenant;
    DELETE FROM warehouses WHERE tenant_id = v_tenant;
    DELETE FROM suppliers WHERE tenant_id = v_tenant;
    DELETE FROM distributors WHERE tenant_id = v_tenant;
    DELETE FROM device_tokens WHERE tenant_id = v_tenant;

    -- Categories
    INSERT INTO categories (id, tenant_id, name, description, level, path)
    VALUES
        (c_fertilizers, v_tenant, 'Fertilizers', 'NPK and crop nutrition', 0, 'Fertilizers'),
        (c_irrigation,  v_tenant, 'Irrigation',  'Drip kits and pumps',   0, 'Irrigation'),
        (c_seeds,       v_tenant, 'Seeds',       'Seed stock for crops',   0, 'Seeds')
    ON CONFLICT (tenant_id, name) DO UPDATE
        SET description = EXCLUDED.description,
            path = EXCLUDED.path;

    -- Warehouses
    INSERT INTO warehouses (id, tenant_id, name, address, capacity, license_number, status)
    VALUES
        (w_main, v_tenant, 'Main Warehouse', 'Central Hub, Sector 21, Mumbai', 12000, 'WH-ACME-01', 'active'),
        (w_east, v_tenant, 'East Warehouse', 'Logistics Park, Pune Bypass',     8000, 'WH-ACME-02', 'active')
    ON CONFLICT (tenant_id, name) DO UPDATE
        SET address = EXCLUDED.address,
            capacity = EXCLUDED.capacity,
            license_number = EXCLUDED.license_number,
            status = EXCLUDED.status;

    -- Suppliers & Distributors
    INSERT INTO suppliers (id, tenant_id, name, contact_email, contact_phone, address, license_number, status)
    VALUES
        (s_agri,  v_tenant, 'Agri Supplies Ltd', 'ops@agrisupplies.example', '+91-98765-43210', 'Plot 9, Industrial Estate, Delhi', 'SUP-ACME-01', 'active'),
        (s_green, v_tenant, 'Green Harvest Co',  'hello@greenharvest.example', '+91-91234-56789', 'NH48, Surat', 'SUP-ACME-02', 'active')
    ON CONFLICT (id) DO UPDATE
        SET name = EXCLUDED.name,
            contact_email = EXCLUDED.contact_email,
            contact_phone = EXCLUDED.contact_phone,
            address = EXCLUDED.address,
            license_number = EXCLUDED.license_number,
            status = EXCLUDED.status;

    INSERT INTO distributors (id, tenant_id, name, contact_email, contact_phone, address, license_number, status)
    VALUES
        (d_farm,   v_tenant, 'Farm Distribution Co', 'sales@farmdistro.example', '+91-90000-00001', 'Kolkata Depot, Ring Road', 'DIST-ACME-01', 'active'),
        (d_coastal,v_tenant, 'Coastal Agro Partners', 'contact@coastalagro.example', '+91-90000-00002', 'Chennai Port Road', 'DIST-ACME-02', 'active')
    ON CONFLICT (id) DO UPDATE
        SET name = EXCLUDED.name,
            contact_email = EXCLUDED.contact_email,
            contact_phone = EXCLUDED.contact_phone,
            address = EXCLUDED.address,
            license_number = EXCLUDED.license_number,
            status = EXCLUDED.status;

    -- Products
    INSERT INTO products (id, tenant_id, category_id, name, quantity, unit_price, unit_of_measure, description, barcode)
    VALUES
        (p_npk,   v_tenant, c_fertilizers, 'NPK 19-19-19 Fertilizer (25kg)', 600, 18.50, 'bag', 'Balanced NPK water soluble fertilizer', 'NPK-191919-25KG'),
        (p_drip,  v_tenant, c_irrigation,  'Drip Irrigation Kit (1 acre)',   180, 32.50, 'kit', 'Complete drip kit with emitters and pipes', 'DRIP-KIT-1AC'),
        (p_seed,  v_tenant, c_seeds,       'Hybrid Basmati Rice Seeds (10kg)', 420, 24.00, 'bag', 'High-yield hybrid basmati seeds', 'SEED-RICE-HYB')
    ON CONFLICT (id) DO UPDATE
        SET category_id = EXCLUDED.category_id,
            name = EXCLUDED.name,
            quantity = EXCLUDED.quantity,
            unit_price = EXCLUDED.unit_price,
            unit_of_measure = EXCLUDED.unit_of_measure,
            description = EXCLUDED.description,
            barcode = EXCLUDED.barcode,
            updated_at = NOW();

    -- Batches
    INSERT INTO batches (id, tenant_id, product_id, batch_number, quantity, expiry_date, manufacturing_date, location, status)
    VALUES
        (b_npk,    v_tenant, p_npk,  'ACME-NPK-2409', 350, CURRENT_DATE + INTERVAL '240 days', CURRENT_DATE - INTERVAL '60 days', 'Aisle A1', 'active'),
        (b_drip_a, v_tenant, p_drip, 'ACME-DRIP-2410', 120, NULL, CURRENT_DATE - INTERVAL '30 days', 'Rack R2', 'active'),
        (b_seed,   v_tenant, p_seed, 'ACME-SEED-2408', 400, CURRENT_DATE + INTERVAL '420 days', CURRENT_DATE - INTERVAL '80 days', 'Cool Storage', 'active')
    ON CONFLICT (product_id, batch_number, tenant_id) DO UPDATE
        SET quantity = EXCLUDED.quantity,
            expiry_date = EXCLUDED.expiry_date,
            manufacturing_date = EXCLUDED.manufacturing_date,
            location = EXCLUDED.location,
            status = EXCLUDED.status,
            updated_at = NOW();

    -- Inventory snapshots
    INSERT INTO inventory (id, tenant_id, warehouse_id, product_id, quantity, reserved_quantity, minimum_level, maximum_level, reorder_point, last_updated)
    VALUES
        (inv_npk_main,  v_tenant, w_main, p_npk,  600, 120, 150, 1200, 200, NOW()),
        (inv_drip_main, v_tenant, w_main, p_drip, 180,  60,  80,  600, 120, NOW()),
        (inv_seed_main, v_tenant, w_main, p_seed, 420,   0, 150, 1000, 220, NOW()),
        (inv_drip_east, v_tenant, w_east, p_drip,  60,  20,  50,  300, 100, NOW())
    ON CONFLICT (tenant_id, warehouse_id, product_id) DO UPDATE
        SET quantity = EXCLUDED.quantity,
            reserved_quantity = EXCLUDED.reserved_quantity,
            minimum_level = EXCLUDED.minimum_level,
            maximum_level = EXCLUDED.maximum_level,
            reorder_point = EXCLUDED.reorder_point,
            last_updated = NOW();

    -- Stock adjustment example (damage)
    INSERT INTO stock_adjustments (id, tenant_id, product_id, warehouse_id, adjustment_type, quantity, previous_stock, new_stock, reason, reference_type, adjusted_by, adjusted_at)
    VALUES
        (adj_drip_damage, v_tenant, p_drip, w_main, 'damage', -30, 240, 210, 'Packaging damage on receipt', 'inspection', v_user, NOW() - INTERVAL '7 days')
    ON CONFLICT (id) DO NOTHING;

    -- Orders (purchase and sales)
    INSERT INTO orders (id, tenant_id, order_type, supplier_id, distributor_id, product_id, warehouse_id, quantity, unit_price, status, order_date, expected_delivery, notes)
    VALUES
        (ord_po_1001, v_tenant, 'purchase', s_agri,    NULL, p_npk,  w_main, 200, 18.50, 'delivered', CURRENT_DATE - INTERVAL '20 days', CURRENT_DATE - INTERVAL '10 days', 'PO-1001 delivered to Main'),
        (ord_po_1002, v_tenant, 'purchase', s_green,   NULL, p_drip, w_main, 300, 19.75, 'processing', CURRENT_DATE - INTERVAL '8 days',  CURRENT_DATE + INTERVAL '2 days', 'PO-1002 in transit'),
        (ord_so_2001, v_tenant, 'sales',   NULL, d_farm,   p_drip, w_main, 120, 32.50, 'shipped',    CURRENT_DATE - INTERVAL '5 days',  CURRENT_DATE + INTERVAL '2 days', 'SO-2001 part shipped'),
        (ord_so_2002, v_tenant, 'sales',   NULL, d_coastal,p_drip, w_east,  30, 34.00, 'pending',    CURRENT_DATE - INTERVAL '2 days',  CURRENT_DATE + INTERVAL '5 days', 'SO-2002 awaiting dispatch'),
        (ord_so_2003, v_tenant, 'sales',   NULL, d_farm,   p_npk,  w_main,  40, 28.00, 'delivered',  CURRENT_DATE - INTERVAL '12 days', CURRENT_DATE - INTERVAL '2 days', 'SO-2003 delivered')
    ON CONFLICT (id) DO UPDATE
        SET supplier_id = EXCLUDED.supplier_id,
            distributor_id = EXCLUDED.distributor_id,
            product_id = EXCLUDED.product_id,
            warehouse_id = EXCLUDED.warehouse_id,
            quantity = EXCLUDED.quantity,
            unit_price = EXCLUDED.unit_price,
            status = EXCLUDED.status,
            order_date = EXCLUDED.order_date,
            expected_delivery = EXCLUDED.expected_delivery,
            notes = EXCLUDED.notes,
            updated_at = NOW();

    -- Order items
    INSERT INTO order_items (id, tenant_id, order_id, product_id, quantity, price)
    VALUES
        (oi_po_1001, v_tenant, ord_po_1001, p_npk,  200, 18.50),
        (oi_po_1002, v_tenant, ord_po_1002, p_drip, 300, 19.75),
        (oi_so_2001, v_tenant, ord_so_2001, p_drip, 120, 32.50),
        (oi_so_2002, v_tenant, ord_so_2002, p_drip,  30, 34.00),
        (oi_so_2003, v_tenant, ord_so_2003, p_npk,   40, 28.00)
    ON CONFLICT (id) DO UPDATE
        SET quantity = EXCLUDED.quantity,
            price = EXCLUDED.price;

    -- Invoices
    INSERT INTO invoices (id, tenant_id, order_id, gstin, hsn_sac, taxable_amount, gst_rate, cgst, sgst, igst, total_amount, status, issued_date, paid_date, invoice_number, due_date)
    VALUES
        (invc_po_1001, v_tenant, ord_po_1001, '27ABCDE1234F2Z5', '3102', 3400.00, 18.0, 306.00, 306.00, 0, 4012.00, 'paid', CURRENT_DATE - INTERVAL '18 days', CURRENT_DATE - INTERVAL '8 days', 'INV-ACME-PO-1001', CURRENT_DATE - INTERVAL '10 days'),
        (invc_so_2003, v_tenant, ord_so_2003, '27ABCDE1234F2Z5', '3102', 1120.00, 12.0, 67.20, 67.20, 0, 1254.40, 'overdue', CURRENT_DATE - INTERVAL '10 days', NULL, 'INV-ACME-SO-2003', CURRENT_DATE - INTERVAL '2 days')
    ON CONFLICT (id) DO UPDATE
        SET status = EXCLUDED.status,
            total_amount = EXCLUDED.total_amount,
            taxable_amount = EXCLUDED.taxable_amount,
            gst_rate = EXCLUDED.gst_rate,
            cgst = EXCLUDED.cgst,
            sgst = EXCLUDED.sgst,
            igst = EXCLUDED.igst,
            issued_date = EXCLUDED.issued_date,
            paid_date = EXCLUDED.paid_date,
            invoice_number = EXCLUDED.invoice_number,
            due_date = EXCLUDED.due_date,
            updated_at = NOW();

    -- Reservations (active hold for pending sales)
    INSERT INTO inventory_reservations (id, tenant_id, product_id, warehouse_id, reservation_id, quantity, reserved_by, reserved_at, expires_at, status, order_id, notes)
    VALUES
        (res_drip_hold, v_tenant, p_drip, w_east, 'SO-2002-RES', 30, v_user, NOW() - INTERVAL '1 day', NOW() + INTERVAL '5 days', 'active', ord_so_2002, 'Reserved for SO-2002 pending dispatch')
    ON CONFLICT (id) DO UPDATE
        SET status = EXCLUDED.status,
            quantity = EXCLUDED.quantity,
            warehouse_id = EXCLUDED.warehouse_id,
            order_id = EXCLUDED.order_id,
            expires_at = EXCLUDED.expires_at,
            notes = EXCLUDED.notes,
            updated_at = NOW();

    -- Device token for push preview
    INSERT INTO device_tokens (id, tenant_id, user_id, device_token, device_type, device_name, app_version, is_active, last_used_at)
    VALUES
        (tok_web, v_tenant, v_user, 'demo-web-token-acme', 'web', 'Acme Web Browser', '1.0.0', true, NOW())
    ON CONFLICT (id) DO UPDATE
        SET device_token = EXCLUDED.device_token,
            device_type = EXCLUDED.device_type,
            device_name = EXCLUDED.device_name,
            app_version = EXCLUDED.app_version,
            is_active = EXCLUDED.is_active,
            last_used_at = NOW(),
            updated_at = NOW();

    -- Notifications for dashboard widgets
    INSERT INTO notifications (id, tenant_id, user_id, title, message, notification_type, priority, status, read_at, created_at)
    VALUES
        (notif_low_stock, v_tenant, v_user, 'Low stock: Drip Kit', 'Inventory for Drip Irrigation Kit is nearing minimum levels in East Warehouse.', 'warning', 'high', 'unread', NULL, NOW() - INTERVAL '1 day'),
        (notif_reorder,   v_tenant, v_user, 'Reorder suggested: NPK 19-19-19', 'Reorder point reached for NPK fertilizer. Consider creating a purchase order.', 'info', 'normal', 'unread', NULL, NOW() - INTERVAL '2 days'),
        (notif_shipment,  v_tenant, v_user, 'Shipment update: SO-2001', 'Sales order SO-2001 has been shipped and is en route.', 'success', 'normal', 'read', NOW() - INTERVAL '1 day', NOW() - INTERVAL '3 days')
    ON CONFLICT (id) DO UPDATE
        SET title = EXCLUDED.title,
            message = EXCLUDED.message,
            notification_type = EXCLUDED.notification_type,
            priority = EXCLUDED.priority,
            status = EXCLUDED.status,
            read_at = EXCLUDED.read_at,
            updated_at = NOW();

    RAISE NOTICE 'Acme seed data applied for tenant %', v_tenant;
END $$;

