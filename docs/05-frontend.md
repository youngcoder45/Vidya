# Phase 5 — Frontend Architecture (Flutter)

**Framework:** Flutter 3.x (Dart 3) · **Minimum:** Android 7+ / iOS 13+ · **Targets:** Android, iOS, tablet (web deferred)

---

## 1. Why Flutter (decision record)

| Criterion | Flutter | React Native (Expo) |
|---|---|---|
| UI consistency | Skia/Impeller renderer — pixel-identical across Android/iOS/tablet; theming via `ThemeData` is first-class | Native widgets → platform divergence; styling via NativeWind/Tailwind adds a layer |
| Performance | AOT-compiled, 60/120 fps; no JS bridge | JS bridge overhead on complex lists (attendance grids, report cards) |
| Developer experience | Single language (Dart), strong static analysis, hot reload | TypeScript familiar; JS ecosystem churn (Expo SDK upgrades, native module pain) |
| Maintainability | `flutter analyze` + strict lints; mono-repo friendly | Large dependency surface; breaking SDK majors |
| Long-term scalability | Google-backed, stable release cadence, excellent tablet support | Facebook/community; native fallbacks needed for edge APIs |
| Cost | One team, one codebase, no per-platform engineers | One codebase but platform quirks still need native work |

**Decision: Flutter** — the ERP's data-dense screens (marks grids, fee ledgers, report cards) demand consistent rendering and smooth list performance; tablet-first admin flows are a Flutter strength. Web (Flutter Web) is deferred — we ship mobile-first per the agreed scope.

---

## 2. Folder Structure

```
mobile/
├── pubspec.yaml
├── analysis_options.yaml
├── lib/
│   ├── main.dart                      # entry, bootstrap providers
│   ├── app.dart                       # MaterialApp.router + theme wiring
│   ├── core/
│   │   ├── config/                    # env config (API base, razorpay key, feature flags)
│   │   ├── network/
│   │   │   ├── api_client.dart        # Dio wrapper: base URL, auth header, refresh-on-401
│   │   │   ├── api_exception.dart     # error envelope → typed exception
│   │   │   └── interceptors.dart      # auth + logging + tenant header
│   │   ├── theme/                     # ⭐ THEME SYSTEM (see §5)
│   │   │   ├── tokens/                # design tokens (colors, typography, spacing, radii)
│   │   │   ├── semantics/             # light/dark semantic color maps
│   │   │   ├── app_theme.dart         # ThemeData factory
│   │   │   └── theme_extensions.dart  # BuildContext extensions (context.colors etc.)
│   │   ├── router/app_router.dart     # go_router: auth guards, roles
│   │   ├── storage/secure_store.dart  # tokens (flutter_secure_storage)
│   │   └── utils/                     # formatters (INR, dates), validators
│   ├── features/
│   │   ├── auth/
│   │   │   ├── data/auth_repository.dart
│   │   │   ├── domain/auth_state.dart
│   │   │   ├── presentation/login_screen.dart
│   │   │   └── presentation/auth_controller.dart
│   │   ├── dashboard/
│   │   │   ├── data/dashboard_repository.dart
│   │   │   ├── domain/dashboard_models.dart
│   │   │   └── presentation/dashboard_screen.dart
│   │   ├── attendance/ …              # future feature modules (same layering)
│   │   ├── homework/ …
│   │   ├── exams/ …
│   │   └── fees/ …
│   └── shared/
│       ├── widgets/                   # app_button, app_text_field, app_card, empty_state,
│       │   │                          # stat_tile, section_header, error_view, shimmer
│       └── widgets/app_scaffold.dart  # branded shell: app bar, nav rail/bottom bar
└── test/                              # widget + unit tests (theme tokens, formatters)
```

**Layering rule:** `presentation → domain ← data`. Features never import each other's internals; cross-feature data (e.g., fees inside student profile) goes through repositories or shared domain models.

---

## 3. State Management — Riverpod

Chosen over Bloc/Provider: compile-safe providers, no boilerplate, testable, first-class async (`AsyncValue`) for network states, and automatic dependency graph (a `dashboardControllerProvider` watching `authControllerProvider` just works).

- **Auth state:** `authControllerProvider` (`Notifier<AuthState>`) — holds user/school/tokens; persists via `SecureStore`; drives router redirects.
- **Server state:** `FutureProvider`/`AsyncNotifier` per screen (`dashboardSummaryProvider`, `studentsProvider(ClassDivisionFilter)`).
- **Mutations:** `Notifier` methods call repositories; invalidate dependent providers (`ref.invalidate(dashboardSummaryProvider)` after fee payment).
- **Theme mode:** `themeModeProvider` (`Notifier<ThemeMode>`) persisted in shared prefs; consumed by `MaterialApp.themeMode` and the `ThemeExt` extensions.
- **Local UI state:** plain `StatefulWidget`/`ValueNotifier` (search text, tab index) — never global.

---

## 4. Navigation — go_router

- URL-style routes: `/login`, `/dashboard`, `/students/:id`, `/fees/ledger/:studentId`, `/exams/:id/results`.
- **Redirect logic** in one place: unauthenticated → `/login`; authenticated without roles → 403 screen; parent routes vs staff routes separated by role branch.
- **Shell route** for the app frame (bottom nav for parents: Dashboard · Homework · Fees · Notices; staff: Dashboard · Students · Attendance · Fees · More) — each role has its own `StatefulShellRoute` with persisted tab state.
- Deep links: notification tap → `pushNotification` payload carries route + params; guarded by auth redirect.
- Dialogs/sheets: pushed as routes (not `showDialog`) so back-navigation and tests are predictable.

---

## 5. Theme System (core deliverable)

### 5.1 Principles
1. **Single source of truth** — every color, spacing, radius, type size lives in `core/theme/tokens/`.
2. **Semantic layer** — widgets reference *roles* (`color: context.colors.primary`), never raw hex.
3. **Zero hardcoded values** — lint rule bans `Color(0x…)`/`Colors.` outside tokens (enforced by `custom_lint` + code review).
4. **Changing one token re-themes the whole app** — `ThemeData` is derived, not duplicated.

### 5.2 Token files

```dart
// core/theme/tokens/app_colors.dart  — RAW PALETTE (the only file with hex values)
abstract final class AppColors {
  static const primary = Color(0xFF1B5E20);        // school brand (overridable per tenant)
  static const primaryContainer = Color(0xFF4C8C4A);
  static const onPrimary = Color(0xFFFFFFFF);
  static const background = Color(0xFFF7F8FA);
  static const surface = Color(0xFFFFFFFF);
  static const surfaceVariant = Color(0xFFEFF1F4);
  static const onSurface = Color(0xFF1C1B1F);
  static const onSurfaceVariant = Color(0xFF44474E);
  static const outline = Color(0xFF74777F);
  static const success = Color(0xFF2E7D32);
  static const warning = Color(0xFFF9A825);
  static const danger = Color(0xFFC62828);
  static const info = Color(0xFF1565C0);
  static const feeAccent = Color(0xFF6A1B9A);      // domain accent
  static const transparent = Color(0x00000000);
}
```

```dart
// core/theme/tokens/app_spacing.dart
abstract final class AppSpacing {
  static const double xxs = 4, xs = 8, sm = 12, md = 16, lg = 24, xl = 32, xxl = 48;
}

// core/theme/tokens/app_typography.dart — type scale (wraps Material TextTheme roles)
abstract final class AppTypography {
  static const displaySmall = TextStyle(fontSize: 32, fontWeight: FontWeight.w700, height: 1.2);
  static const headline = TextStyle(fontSize: 24, fontWeight: FontWeight.w700, height: 1.25);
  static const title = TextStyle(fontSize: 18, fontWeight: FontWeight.w600, height: 1.3);
  static const body = TextStyle(fontSize: 14, fontWeight: FontWeight.w400, height: 1.5);
  static const bodyStrong = TextStyle(fontSize: 14, fontWeight: FontWeight.w600, height: 1.5);
  static const label = TextStyle(fontSize: 12, fontWeight: FontWeight.w500, height: 1.4);
  static const caption = TextStyle(fontSize: 11, fontWeight: FontWeight.w400, height: 1.3);
}

// core/theme/tokens/app_radii.dart
abstract final class AppRadii {
  static const double xs = 6, sm = 10, md = 14, lg = 20, pill = 999;
}
```

### 5.3 Semantic layer

```dart
// core/theme/semantics/app_semantics.dart
class AppSemantics {
  const AppSemantics({
    required this.primary, required this.onPrimary, required this.surface,
    required this.background, required this.text, required this.textMuted,
    required this.border, required this.success, required this.warning,
    required this.danger, required this.info, required this.feeAccent,
    required this.surfaceVariant, required this.scrim,
  });
  final Color primary, onPrimary, surface, background, text, textMuted, border;
  final Color success, warning, danger, info, feeAccent, surfaceVariant, scrim;

  static const light = AppSemantics(
    primary: AppColors.primary, onPrimary: AppColors.onPrimary,
    surface: AppColors.surface, background: AppColors.background,
    text: AppColors.onSurface, textMuted: AppColors.onSurfaceVariant,
    border: AppColors.outline, success: AppColors.success,
    warning: AppColors.warning, danger: AppColors.danger, info: AppColors.info,
    feeAccent: AppColors.feeAccent, surfaceVariant: AppColors.surfaceVariant,
    scrim: Color(0x66000000),
  );

  static const dark = AppSemantics(
    primary: Color(0xFF81C784),          // lightened for contrast on dark
    onPrimary: Color(0xFF003300),
    surface: Color(0xFF1E1E24), background: Color(0xFF141418),
    text: Color(0xFFE6E1E5), textMuted: Color(0xFFA8AAB3),
    border: Color(0xFF5A5D66), success: Color(0xFF81C784),
    warning: Color(0xFFFFD54F), danger: Color(0xFFEF9A9A), info: Color(0xFF90CAF9),
    feeAccent: Color(0xFFCE93D8), surfaceVariant: Color(0xFF2A2A31),
    scrim: Color(0xB3000000),
  );
}
```

### 5.4 ThemeData factory

```dart
// core/theme/app_theme.dart
ThemeData buildAppTheme(AppSemantics s, Brightness brightness) {
  final scheme = ColorScheme(
    brightness: brightness,
    primary: s.primary, onPrimary: s.onPrimary,
    surface: s.surface, onSurface: s.text,
    surfaceContainerHighest: s.surfaceVariant, onSurfaceVariant: s.textMuted,
    outline: s.border, error: s.danger, onError: Colors.white,
    secondary: s.info, onSecondary: s.surface,
  );
  return ThemeData(
    useMaterial3: true,
    colorScheme: scheme,
    scaffoldBackgroundColor: s.background,
    textTheme: TextTheme(
      displaySmall: AppTypography.displaySmall.copyWith(color: s.text),
      headlineMedium: AppTypography.headline.copyWith(color: s.text),
      titleLarge: AppTypography.title.copyWith(color: s.text),
      bodyMedium: AppTypography.body.copyWith(color: s.text),
      bodyLarge: AppTypography.bodyStrong.copyWith(color: s.text),
      labelMedium: AppTypography.label.copyWith(color: s.textMuted),
      bodySmall: AppTypography.caption.copyWith(color: s.textMuted),
    ),
    appBarTheme: AppBarTheme(backgroundColor: s.surface, foregroundColor: s.text, elevation: 0),
    cardTheme: CardThemeData(color: s.surface, elevation: 0, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadii.md), side: BorderSide(color: s.border.withValues(alpha: .2)))),
    inputDecorationTheme: InputDecorationTheme(filled: true, fillColor: s.surfaceVariant, border: OutlineInputBorder(borderRadius: BorderRadius.circular(AppRadii.sm), borderSide: BorderSide(color: s.border))),
    dividerTheme: DividerThemeData(color: s.border.withValues(alpha: .25)),
    snackBarTheme: SnackBarThemeData(behavior: SnackBarBehavior.floating, backgroundColor: s.text),
    elevatedButtonTheme: ElevatedButtonThemeData(style: ElevatedButton.styleFrom(backgroundColor: s.primary, foregroundColor: s.onPrimary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadii.sm)), minimumSize: const Size.fromHeight(48))),
    extensions: [s],                                  // AppSemantics available via context
  );
}

final appLightTheme = buildAppTheme(AppSemantics.light, Brightness.light);
final appDarkTheme = buildAppTheme(AppSemantics.dark, Brightness.dark);
```

### 5.5 Context extensions — the only way widgets touch theme

```dart
// core/theme/theme_extensions.dart
extension ThemeExt on BuildContext {
  AppSemantics get colors => Theme.of(this).extension<AppSemantics>()!;
  TextTheme get typography => Theme.of(this).textTheme;
  double get spaceXxs => AppSpacing.xxs;  // + xs, sm, md, lg, xl, xxl
  ColorScheme get scheme => Theme.of(this).colorScheme;
  bool get isDark => Theme.of(this).brightness == Brightness.dark;
}
```

### 5.6 Usage rule (enforced)

```dart
// ✅ correct — semantic + tokens only
Card(
  color: context.colors.surface,
  child: Padding(
    padding: EdgeInsets.all(context.spaceMd),
    child: Text('₹12,500 due', style: context.typography.titleLarge?.copyWith(color: context.colors.danger)),
  ),
)
// ❌ forbidden — hardcoded: Color(0xFF…), Colors.grey, EdgeInsets.all(16), fontSize: 14
```

### 5.7 Per-tenant branding hook

School branding (`branding.primary_color` from the tenant API) overrides `AppColors.primary` **at token load time** — a `brandedThemeProvider` rebuilds the `AppSemantics` with the school's color while keeping every other token. One variable changes the whole school's app.

---

## 6. Accessibility

- Semantic contrast: all text pairs meet WCAG AA (validated by token pairs, not ad-hoc colors).
- `MediaQuery.textScaler` respected (no fixed `fontSize` overrides outside tokens).
- Min 44 px touch targets (`Size.fromHeight(48)` buttons, padded list tiles).
- Screen-reader labels on icon-only actions; semantic labels on charts.
- Dark mode is a first-class token set, not an afterthought (`themeMode` = system default, user-overridable).

## 7. API client behavior

- Dio + interceptor chain: attach `Authorization`; on 401 → single-flight refresh (queue concurrent calls), retry once; on 403/422/429 → typed `ApiException` mapped to user-friendly message via error codes; tenant `X-School-ID` header on pre-auth calls.
- Timeouts: connect 10 s, receive 20 s; retry with backoff for idempotent GETs only.
- Money shown as `₹1,23,450` (Indian grouping) via a shared formatter; amounts always `int` paise from API.
