import 'package:flutter/material.dart';

import 'semantics/app_semantics.dart';
import 'tokens/app_spacing.dart';
import 'tokens/app_typography.dart';

/// The only sanctioned way widgets consume the theme.
///
/// Usage rules (enforced in review/CI):
///  * colors  → `context.colors.<role>`          (never raw `Color(0x…)`)
///  * type    → `context.typography.<role>`      (never `fontSize: 14`)
///  * spacing → `context.spaceMd` etc.           (never `EdgeInsets.all(16)`)
///  * shapes  → via ThemeData (tokens already wired)
extension ThemeExt on BuildContext {
  /// Semantic color roles for the current brightness.
  AppSemantics get colors => Theme.of(this).extension<AppSemantics>()!;

  /// Type roles mapped to the current theme.
  TextTheme get typography => Theme.of(this).textTheme;

  /// Spacing scale (direct token access is allowed for layout).
  double get spaceXxs => AppSpacing.xxs;
  double get spaceXs => AppSpacing.xs;
  double get spaceSm => AppSpacing.sm;
  double get spaceMd => AppSpacing.md;
  double get spaceLg => AppSpacing.lg;
  double get spaceXl => AppSpacing.xl;
  double get spaceXxl => AppSpacing.xxl;

  /// The Material color scheme (component-level theming).
  ColorScheme get scheme => Theme.of(this).colorScheme;

  /// Whether the app is currently in dark mode.
  bool get isDark => Theme.of(this).brightness == Brightness.dark;
}
