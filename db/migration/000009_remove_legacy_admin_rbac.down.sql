CREATE TABLE "roles" (
                         "id" bigserial PRIMARY KEY,
                         "role_name" varchar UNIQUE NOT NULL,
                         "description" text NOT NULL DEFAULT '',
                         "is_system" bool NOT NULL DEFAULT false,
                         "created_at" timestamptz NOT NULL DEFAULT (now())
);

INSERT INTO roles (role_name, description, is_system)
VALUES ('super', '系统内置超级管理员，拥有全部权限', true);

INSERT INTO roles (role_name, description, is_system)
VALUES ('admin', '普通管理员，拥有基本管理权限', true);

CREATE TABLE "admins" (
                          "id" bigserial PRIMARY KEY,
                          "username" varchar UNIQUE NOT NULL,
                          "hashed_password" varchar NOT NULL,
                          "is_active" bool NOT NULL DEFAULT false,
                          "role_id" bigint NOT NULL DEFAULT 2,
                          "created_at" timestamptz NOT NULL DEFAULT (now()),
                          "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

INSERT INTO admins (username, hashed_password, is_active, role_id)
VALUES (
           'super',
           '$2a$10$pXJDlyeVCvQrGu.sPsxS4.I6xgPM1ewS1K8pjfo9OTKYs4VZxYp62',
           true,
           1
       );

CREATE TABLE "sys_menus" (
                             "id" bigserial PRIMARY KEY,
                             "name" varchar NOT NULL,
                             "path" varchar NOT NULL DEFAULT '',
                             "icon" varchar NOT NULL DEFAULT '',
                             "is_active" bool NOT NULL DEFAULT false,
                             "type" int NOT NULL DEFAULT 2,
                             "sort" int NOT NULL DEFAULT 0,
                             "parent_id" bigint DEFAULT null,
                             "created_at" timestamptz NOT NULL DEFAULT (now()),
                             "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

INSERT INTO sys_menus (name, path, icon, is_active, type, sort)
VALUES ('基本管理', '', 'pi pi-sliders-h', true, 1, 1);

INSERT INTO sys_menus (name, path, icon, is_active, type, parent_id, sort)
VALUES ('文章管理', '/manage/articles', 'pi pi-book', true, 2, 1, 2);

INSERT INTO sys_menus (id, name, path, icon, is_active, type, parent_id, sort)
VALUES (3, '分类管理', '/manage/categories', 'pi pi-tag', true, 2, 1, 3);

CREATE TABLE "role_permissions" (
                                    "id" bigserial PRIMARY KEY,
                                    "role_id" bigint NOT NULL,
                                    "menu_id" bigint NOT NULL,
                                    "created_by" bigint NOT NULL,
                                    "created_at" timestamptz NOT NULL DEFAULT (now())
);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 1, 1);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 2, 1);

INSERT INTO role_permissions (role_id, menu_id, created_by)
VALUES (1, 3, 1);

CREATE UNIQUE INDEX ON "sys_menus" ("name", "parent_id");

CREATE UNIQUE INDEX ON "role_permissions" ("role_id", "menu_id");

ALTER TABLE "admins" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "sys_menus" ADD FOREIGN KEY ("parent_id") REFERENCES "sys_menus" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("menu_id") REFERENCES "sys_menus" ("id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("created_by") REFERENCES "admins" ("id");
