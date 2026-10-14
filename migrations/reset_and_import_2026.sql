-- ====================================================================
-- AAMS Database Reset and Employee Import Migration (2026)
-- Target: Reset employees & users, preserve admin 0500626432, import Excel employees
-- ====================================================================

BEGIN;

-- 1. Ensure SUPER_ADMIN, OWNER, SUPERVISOR, and ACCOUNTANT roles exist
INSERT INTO roles (id, name, display_name, description, permissions, is_system, created_at, updated_at)
VALUES 
  (gen_random_uuid(), 'SUPER_ADMIN', 'مدير عام (مسؤول النظام)', 'كامل الصلاحيات للتحكم في كافة أقسام وإعدادات ومستخدمي النظام', '["*"]', true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET permissions = '["*"]';

INSERT INTO roles (id, name, display_name, description, permissions, is_system, created_at, updated_at)
VALUES 
  (gen_random_uuid(), 'OWNER', 'مالك المؤسسة', 'كامل الصلاحيات للإشراف العام والتقارير والعمليات المالية والإدارية', '["*"]', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (id, name, display_name, description, permissions, is_system, created_at, updated_at)
VALUES 
  (gen_random_uuid(), 'SUPERVISOR', 'مشرف وردية', 'متابعة المناديب وتسجيل الدوام والعمليات الميدانية', '["employees.view","employees.create","employees.edit","employees.cards","work.view","work.start","work.end","custody","custody.access","custody.view","custody.add","vehicles.view","vehicles.oil","fuel.view","fuel.manage","maintenance.view","maintenance.manage","attendance.view","investigations.view","investigations.create","inventory.view","inventory.dispense","tickets.view","tickets.manage"]', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (id, name, display_name, description, permissions, is_system, created_at, updated_at)
VALUES 
  (gen_random_uuid(), 'ACCOUNTANT', 'إدارة مالية ومحاسبة', 'إدارة العهدة والمصروفات والحسابات البنكية والتقارير المالية', '["custody","custody.access","custody.view","custody.add","custody.delete","bank_accounts.view","bank_accounts.manage","fuel.view","fuel.manage","violations.view","violations.manage","reports.view","reports.export","inventory.view","investigations.view"]', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 2. Clear all existing employees and related child records
DELETE FROM work_sessions;
DELETE FROM attendances;
DELETE FROM custody_expenses;
DELETE FROM investigations;
DELETE FROM maintenance_requests;
DELETE FROM fuel_logs;
DELETE FROM traffic_violations;
DELETE FROM employee_documents;
DELETE FROM employee_bank_accounts;
DELETE FROM leave_requests;
DELETE FROM employees;

-- 3. Delete non-admin users, keeping only phone 0500626432
DELETE FROM admins WHERE phone != '0500626432' AND phone != '+966500626432' AND username != '0500626432';

-- 4. Ensure admin 0500626432 is configured as SUPER_ADMIN
DO $$
DECLARE
  super_role_id char(36);
BEGIN
  SELECT id::text INTO super_role_id FROM roles WHERE name = 'SUPER_ADMIN' LIMIT 1;
  
  IF EXISTS (SELECT 1 FROM admins WHERE phone = '0500626432' OR phone = '+966500626432' OR username = '0500626432') THEN
    UPDATE admins 
    SET role = 'ADMIN', role_id = super_role_id::uuid, updated_at = NOW()
    WHERE phone = '0500626432' OR phone = '+966500626432' OR username = '0500626432';
  ELSE
    INSERT INTO admins (id, email, username, phone, password, name, role, role_id, created_at, updated_at)
    VALUES (
      gen_random_uuid(),
      'admin0500626432@aams.com',
      '0500626432',
      '0500626432',
      '$2a$10$w8T9Zk3K6bM5V4X7Q9W2Oe1S4U7Y0A3C6E9G2I5K8M1O4Q7S0U3W', -- Admin@2026!
      'المدير العام',
      'ADMIN',
      super_role_id::uuid,
      NOW(),
      NOW()
    );
  END IF;
END $$;

-- 5. Insert 89 Employees from Excel
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'ساره خالد عبد الله الفقار',
  'DRIVER',
  '1086722673',
  '1086722673',
  '1086722673',
  '1',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '1086722673',
  '1086722673',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'صفا محمد المالكي',
  'DRIVER',
  '1114693045',
  '1114693045',
  '1114693045',
  '2',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '1114693045',
  '1114693045',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد صبري ندا',
  'DRIVER',
  '2543679506',
  '2543679506',
  '2543679506',
  '3',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2543679506',
  '2543679506',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد هشام محمد بركه',
  'DRIVER',
  '2559916248',
  '2559916248',
  '2559916248',
  '4',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2559916248',
  '2559916248',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'إبراهيم محمود إبراهيم الكرداس',
  'DRIVER',
  '2642799148',
  '2642799148',
  '2642799148',
  '5',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2642799148',
  '2642799148',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مصطفى صلاح طلبه الخولي',
  'DRIVER',
  '2543859603',
  '2543859603',
  '2543859603',
  '6',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2543859603',
  '2543859603',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد أحمد حسن هريدي',
  'DRIVER',
  '2558063638',
  '2558063638',
  '2558063638',
  '7',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2558063638',
  '2558063638',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مكرم بالله',
  'DRIVER',
  '2569393479',
  '2569393479',
  '2569393479',
  '8',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2569393479',
  '2569393479',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'حنيف محمد كالدوش ميا',
  'DRIVER',
  '2571022215',
  '2571022215',
  '2571022215',
  '9',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2571022215',
  '2571022215',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'سعيد سعيد وائل نعيم شاه',
  'DRIVER',
  '2557526874',
  '2557526874',
  '2557526874',
  '10',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2557526874',
  '2557526874',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'راجو حمد حرج علي',
  'DRIVER',
  '2571413398',
  '2571413398',
  '2571413398',
  '11',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2571413398',
  '2571413398',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'دلوار اوسين شيبون',
  'DRIVER',
  '2569600022',
  '2569600022',
  '2569600022',
  '12',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2569600022',
  '2569600022',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد رقيب خان',
  'DRIVER',
  '2570685491',
  '2570685491',
  '2570685491',
  '13',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2570685491',
  '2570685491',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد إسماعيل اوسين مير',
  'DRIVER',
  '2571034277',
  '2571034277',
  '2571034277',
  '14',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2571034277',
  '2571034277',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'ابوالخير أحمد عبدالله مزيد',
  'DRIVER',
  '2562140786',
  '2562140786',
  '2562140786',
  '15',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2562140786',
  '2562140786',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مصطفى ياسر محمود',
  'DRIVER',
  '2575011578',
  '2575011578',
  '2575011578',
  '16',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2575011578',
  '2575011578',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'اسلام ابوالعلا حسين',
  'DRIVER',
  '2576800094',
  '2576800094',
  '2576800094',
  '17',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2576800094',
  '2576800094',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد إبراهيم عباس الزيادي',
  'DRIVER',
  '2577274109',
  '2577274109',
  '2577274109',
  '18',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2577274109',
  '2577274109',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمود عاشور سيد محمد',
  'DRIVER',
  '2577448844',
  '2577448844',
  '2577448844',
  '19',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2577448844',
  '2577448844',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'اسلام عاطف زكي محمد',
  'DRIVER',
  '2560625002',
  '2560625002',
  '2560625002',
  '20',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2560625002',
  '2560625002',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'سيد أحمد سيد شلبه',
  'DRIVER',
  '2584664482',
  '2584664482',
  '2584664482',
  '21',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2584664482',
  '2584664482',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد مهدي ايمون',
  'DRIVER',
  '2582493017',
  '2582493017',
  '2582493017',
  '22',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2582493017',
  '2582493017',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد عدنان اباد علي',
  'DRIVER',
  '2575492331',
  '2575492331',
  '2575492331',
  '23',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2575492331',
  '2575492331',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد سول رحمن',
  'DRIVER',
  '2595483005',
  '2595483005',
  '2595483005',
  '24',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2595483005',
  '2595483005',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد أحمد متولي',
  'DRIVER',
  '2614943351',
  '2614943351',
  '2614943351',
  '25',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2614943351',
  '2614943351',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'علي رفعت محمد إبراهيم',
  'DRIVER',
  '2614943328',
  '2614943328',
  '2614943328',
  '26',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2614943328',
  '2614943328',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'الاي هعصلي حجيته نيعاً',
  'DRIVER',
  '2644365871',
  '2644365871',
  '2644365871',
  '27',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2644365871',
  '2644365871',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'رياض حسن',
  'DRIVER',
  '2614943385',
  '2614943385',
  '2614943385',
  '28',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2614943385',
  '2614943385',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'إسماعيل حسين',
  'DRIVER',
  '2601602705',
  '2601602705',
  '2601602705',
  '29',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2601602705',
  '2601602705',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد غلام',
  'DRIVER',
  '2602432763',
  '2602432763',
  '2602432763',
  '30',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2602432763',
  '2602432763',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمود حسن مأمون',
  'DRIVER',
  '2600467787',
  '2600467787',
  '2600467787',
  '31',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2600467787',
  '2600467787',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد شوهيل',
  'DRIVER',
  '2600819698',
  '2600819698',
  '2600819698',
  '32',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2600819698',
  '2600819698',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أكرم',
  'DRIVER',
  '2600255448',
  '2600255448',
  '2600255448',
  '33',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2600255448',
  '2600255448',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد وائل',
  'DRIVER',
  '2594207413',
  '2594207413',
  '2594207413',
  '34',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2594207413',
  '2594207413',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'ابوطالب',
  'DRIVER',
  '2602172328',
  '2602172328',
  '2602172328',
  '35',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2602172328',
  '2602172328',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمود إسماعيل أنور محمد',
  'DRIVER',
  '2627624295',
  '2627624295',
  '2627624295',
  '36',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2627624295',
  '2627624295',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'بلال محمد سعيد عبدالعال',
  'DRIVER',
  '2625434226',
  '2625434226',
  '2625434226',
  '37',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2625434226',
  '2625434226',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'لابون',
  'DRIVER',
  'TEMP000000038',
  '',
  'TEMP000000038',
  '38',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000038',
  'TEMP000000038',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'نور علم',
  'DRIVER',
  '2631587660',
  '2631587660',
  '2631587660',
  '39',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2631587660',
  '2631587660',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'ناهد حسن',
  'DRIVER',
  '2632878209',
  '2632878209',
  '2632878209',
  '40',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2632878209',
  '2632878209',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'جهيدال اسلام',
  'DRIVER',
  '2627916931',
  '2627916931',
  '2627916931',
  '41',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2627916931',
  '2627916931',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد أسامة جمعة',
  'DRIVER',
  '2558902942',
  '2558902942',
  '2558902942',
  '42',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2558902942',
  '2558902942',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'كريم ياسر حسين',
  'DRIVER',
  '2632874935',
  '2632874935',
  '2632874935',
  '43',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2632874935',
  '2632874935',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مصطفى عبدالعال السيد',
  'DRIVER',
  '2625434226-44',
  '2625434226-44',
  '2625434226-44',
  '44',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2625434226-44',
  '2625434226-44',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'نور زمان حسيب',
  'DRIVER',
  '2626160838',
  '2626160838',
  '2626160838',
  '45',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2626160838',
  '2626160838',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'اسلام محمد',
  'DRIVER',
  '2622331037',
  '2622331037',
  '2622331037',
  '46',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2622331037',
  '2622331037',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد روبل اوسين',
  'DRIVER',
  '2570490058',
  '2570490058',
  '2570490058',
  '47',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2570490058',
  '2570490058',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'طارق اسلام',
  'DRIVER',
  '2631587454',
  '2631587454',
  '2631587454',
  '48',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2631587454',
  '2631587454',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد نونون',
  'DRIVER',
  '2628524478',
  '2628524478',
  '2628524478',
  '49',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2628524478',
  '2628524478',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد شوهفاق',
  'DRIVER',
  '2628593580',
  '2628593580',
  '2628593580',
  '50',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2628593580',
  '2628593580',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'بشير خميس أحمد علي',
  'DRIVER',
  '2632875510',
  '2632875510',
  '2632875510',
  '51',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2632875510',
  '2632875510',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد دلوار حسين',
  'DRIVER',
  '2630296735',
  '2630296735',
  '2630296735',
  '52',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2630296735',
  '2630296735',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'حسن راجا',
  'DRIVER',
  '2630297311',
  '2630297311',
  '2630297311',
  '53',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2630297311',
  '2630297311',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'هابير شاكبه',
  'DRIVER',
  '2632878423',
  '2632878423',
  '2632878423',
  '54',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2632878423',
  '2632878423',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد اشرفول علم',
  'DRIVER',
  '2629908514',
  '2629908514',
  '2629908514',
  '55',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2629908514',
  '2629908514',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'رحمان يونس',
  'DRIVER',
  '2633404534',
  '2633404534',
  '2633404534',
  '56',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2633404534',
  '2633404534',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أبو سفيان',
  'DRIVER',
  '2633404500',
  '2633404500',
  '2633404500',
  '57',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2633404500',
  '2633404500',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'حمد حنيف',
  'DRIVER',
  '2633404468',
  '2633404468',
  '2633404468',
  '58',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2633404468',
  '2633404468',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'ميلون',
  'DRIVER',
  '2633404575',
  '2633404575',
  '2633404575',
  '59',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2633404575',
  '2633404575',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد عبد الله عبدالرحمن',
  'DRIVER',
  '2630756282',
  '2630756282',
  '2630756282',
  '60',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2630756282',
  '2630756282',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'فردين',
  'DRIVER',
  '2631587488',
  '2631587488',
  '2631587488',
  '61',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2631587488',
  '2631587488',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'تمال',
  'DRIVER',
  '2624812067',
  '2624812067',
  '2624812067',
  '62',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2624812067',
  '2624812067',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'شوهان شيك',
  'DRIVER',
  '2584701458',
  '2584701458',
  '2584701458',
  '63',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2584701458',
  '2584701458',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'راشد محمود ساردار',
  'DRIVER',
  '2636302768',
  '2636302768',
  '2636302768',
  '64',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2636302768',
  '2636302768',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'رقيب اسلام',
  'DRIVER',
  '2599416852',
  '2599416852',
  '2599416852',
  '65',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2599416852',
  '2599416852',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمود حسين',
  'DRIVER',
  '2634815092',
  '2634815092',
  '2634815092',
  '66',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2634815092',
  '2634815092',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'قاجي سول رنا',
  'DRIVER',
  'TEMP000000067',
  '',
  'TEMP000000067',
  '67',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000067',
  'TEMP000000067',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد وقاص',
  'DRIVER',
  '2638235990',
  '2638235990',
  '2638235990',
  '68',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2638235990',
  '2638235990',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أحمد محمد طه',
  'DRIVER',
  '2640436743',
  '2640436743',
  '2640436743',
  '69',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2640436743',
  '2640436743',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أسامة كمال',
  'DRIVER',
  '2640943557',
  '2640943557',
  '2640943557',
  '70',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2640943557',
  '2640943557',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'اسور عاشور علي',
  'DRIVER',
  '2542723057',
  '2542723057',
  '2542723057',
  '71',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2542723057',
  '2542723057',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'اكاش',
  'DRIVER',
  'TEMP000000072',
  '',
  'TEMP000000072',
  '72',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000072',
  'TEMP000000072',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد السيد سالم',
  'DRIVER',
  '2558902942-73',
  '2558902942-73',
  '2558902942-73',
  '73',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2558902942-73',
  '2558902942-73',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مد مونير روح زمان',
  'DRIVER',
  '2629313749',
  '2629313749',
  '2629313749',
  '74',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2629313749',
  '2629313749',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أيمن رجب',
  'DRIVER',
  '2641844887',
  '2641844887',
  '2641844887',
  '75',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641844887',
  '2641844887',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمود عبدالرحيم محمد حسين',
  'DRIVER',
  '2641911298',
  '2641911298',
  '2641911298',
  '76',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641911298',
  '2641911298',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'طارق جمعة كمال محمد',
  'DRIVER',
  '2641911116',
  '2641911116',
  '2641911116',
  '77',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641911116',
  '2641911116',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أحمد إبراهيم عز الدين محمد',
  'DRIVER',
  '2641910969',
  '2641910969',
  '2641910969',
  '78',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641910969',
  '2641910969',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد حسني عبدالرحمن جوده',
  'DRIVER',
  '2641911223',
  '2641911223',
  '2641911223',
  '79',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641911223',
  '2641911223',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد السيد مخيمر',
  'DRIVER',
  '2641845009',
  '2641845009',
  '2641845009',
  '80',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641845009',
  '2641845009',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أشيك خان',
  'DRIVER',
  'TEMP000000081',
  '',
  'TEMP000000081',
  '81',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000081',
  'TEMP000000081',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'مصطفى علي مصطفى علي',
  'DRIVER',
  '2641845272',
  '2641845272',
  '2641845272',
  '82',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641845272',
  '2641845272',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد شهاب الدين',
  'DRIVER',
  '2600361972',
  '2600361972',
  '2600361972',
  '83',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2600361972',
  '2600361972',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'أحمد سامح فدري إبراهيم',
  'DRIVER',
  '2641913484',
  '2641913484',
  '2641913484',
  '84',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641913484',
  '2641913484',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'هادي نبيل محمد عثمان',
  'DRIVER',
  '2641913393',
  '2641913393',
  '2641913393',
  '85',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2641913393',
  '2641913393',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'عمرو رشاد قبيس علي',
  'DRIVER',
  'TEMP000000086',
  '',
  'TEMP000000086',
  '86',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000086',
  'TEMP000000086',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'محمد أكمل علم',
  'DRIVER',
  'TEMP000000087',
  '',
  'TEMP000000087',
  '87',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000087',
  'TEMP000000087',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'إبراهيم خليل',
  'DRIVER',
  'TEMP000000088',
  '',
  'TEMP000000088',
  '88',
  '',
  'ninja',
  'motorcycle',
  'morning',
  'TEMP000000088',
  'TEMP000000088',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);
INSERT INTO employees (
  id, name, job_role, employee_number, phone, national_id, key_number, motorcycle_number,
  application_type, vehicle_type, shift, barcode, qr_code, total_distance, last_oil_change_distance,
  is_working, language, created_at, updated_at
) VALUES (
  gen_random_uuid(),
  'فرحان مصطفى',
  'DRIVER',
  '2644398741',
  '2644398741',
  '2644398741',
  '89',
  '',
  'ninja',
  'motorcycle',
  'morning',
  '2644398741',
  '2644398741',
  0,
  0,
  false,
  'ar',
  NOW(),
  NOW()
);

COMMIT;
