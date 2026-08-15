import 'package:flutter/material.dart';

/// RAW PALETTE — the single source of truth for every color in the app.
///
/// Rules:
///  * This is the ONLY file where `Color(0x…)` literals are allowed.
///  * Widgets must consume colors through the semantic layer
///    (`context.colors.*`) defined in `core/theme/semantics/`.
///  * Changing a token here re-themes the entire application.
///
/// The [primary] family is the default school brand. When a tenant provides
/// its own branding (see docs/05 §5.7), [AppColors.primary] is overridden at
/// token-load time with the school's color — every other token stays intact.
abstract final class AppColors {
  // Brand
  static const Color primary = Color(0xFF1B5E20);
  static const Color primaryContainer = Color(0xFF4C8C4A);
  static const Color onPrimary = Color(0xFFFFFFFF);
  static const Color primaryDark = Color(0xFF81C784);

  // Surfaces & backgrounds
  static const Color background = Color(0xFFF7F8FA);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceVariant = Color(0xFFEFF1F4);
  static const Color scrim = Color(0x66000000);

  // Text & borders
  static const Color onSurface = Color(0xFF1C1B1F);
  static const Color onSurfaceVariant = Color(0xFF44474E);
  static const Color outline = Color(0xFF74777F);

  // Semantic status
  static const Color success = Color(0xFF2E7D32);
  static const Color warning = Color(0xFFF9A825);
  static const Color danger = Color(0xFFC62828);
  static const Color info = Color(0xFF1565C0);

  // Domain accents
  static const Color feeAccent = Color(0xFF6A1B9A);

  // Dark-mode overrides (semantic mapping in AppSemantics.dark)
  static const Color darkBackground = Color(0xFF141418);
  static const Color darkSurface = Color(0xFF1E1E24);
  static const Color darkSurfaceVariant = Color(0xFF2A2A31);
  static const Color darkOnSurface = Color(0xFFE6E1E5);
  static const Color darkOnSurfaceVariant = Color(0xFFA8AAB3);
  static const Color darkOutline = Color(0xFF5A5D66);
  static const Color darkScrim = Color(0xB3000000);
}
