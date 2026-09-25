# SchoolOS — Multi-Tenant School ERP SaaS

> A production-oriented, multi-tenant School ERP SaaS built with **Flutter, Go, PostgreSQL, and Redis**.

SchoolOS is a mobile-first school management platform designed around a scalable **multi-tenant architecture**, with support for students, attendance, fees, announcements, notifications, dashboards, RBAC, and online payments.

<p align="center">
  <img src="https://img.shields.io/badge/Flutter-3.x-02569B?logo=flutter&logoColor=white" alt="Flutter">
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Redis-7+-DC382D?logo=redis&logoColor=white" alt="Redis">
  <img src="https://img.shields.io/badge/Razorpay-Payments-528FF0" alt="Razorpay">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white" alt="Docker">
</p>


---

## Overview

This repository contains:

- Complete architecture documentation across **8 phases**
- A compiling **Go backend scaffold**
- A **Flutter mobile application**
- PostgreSQL database schema and migrations
- Redis integration
- Multi-tenant data isolation
- JWT authentication and RBAC
- Razorpay payment integration
- Notifications and announcements
- Dashboard infrastructure
- Docker-based local infrastructure
- CI/CD configuration

The current scaffold implements the core School ERP functionality while keeping the architecture ready for additional modules.

### Currently Implemented

- Authentication
- Multi-tenant school setup
- Students
- Attendance
- Fees
- Offline payments
- Razorpay orders and webhooks
- Announcements
- Notifications
- Dashboard
- RBAC
- Tenant isolation
- Light/dark theme system

### Designed but Not Yet Implemented

The following modules are fully designed and schema'd, but currently registered as `501 Not Implemented` stubs:

- Homework
- Examinations
- Payroll

---

## New Here?

If you're new to the project, start with **[`SETUP.md`](SETUP.md)**.

It contains a beginner-friendly, from-scratch setup guide covering the entire stack and getting the project running locally.

---
