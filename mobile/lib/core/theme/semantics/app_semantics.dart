import 'package:flutter/material.dart';

import '../tokens/app_colors.dart';

/// Semantic color roles — what widgets actually reference.
///
/// Maps raw tokens ([AppColors]) to intent-driven roles. Light and dark
/// variants live here so a widget written once works in both modes without
/// branching on brightness. Swap [light]/[dark] or provide a tenant-branded
/// instance to re-theme the whole app.
class AppSemantics extends ThemeExtension<AppSemantics> {
  const AppSemantics({
    required this.primary,
    required this.onPrimary,
    required this.surface,
    required this.surfaceVariant,
    required this.background,
    required this.text,
    required this.textMuted,
    required this.border,
    required this.success,
    required this.warning,
    required this.danger,
    required this.info,
    required this.feeAccent,
    required this.scrim,
  });

  final Color primary;
  final Color onPrimary;
  final Color surface;
  final Color surfaceVariant;
  final Color background;
  final Color text;
  final Color textMuted;
  final Color border;
  final Color success;
  final Color warning;
  final Color danger;
  final Color info;
  final Color feeAccent;
  final Color scrim;

  static const AppSemantics light = AppSemantics(
    primary: AppColors.primary,
    onPrimary: AppColors.onPrimary,
    surface: AppColors.surface,
    surfaceVariant: AppColors.surfaceVariant,
    background: AppColors.background,
    text: AppColors.onSurface,
    textMuted: AppColors.onSurfaceVariant,
    border: AppColors.outline,
    success: AppColors.success,
    warning: AppColors.warning,
    danger: AppColors.danger,
    info: AppColors.info,
    feeAccent: AppColors.feeAccent,
    scrim: AppColors.scrim,
  );

  static const AppSemantics dark = AppSemantics(
    primary: AppColors.primaryDark,
    onPrimary: Color(0xFF003300),
    surface: AppColors.darkSurface,
    surfaceVariant: AppColors.darkSurfaceVariant,
    background: AppColors.darkBackground,
    text: AppColors.darkOnSurface,
    textMuted: AppColors.darkOnSurfaceVariant,
    border: AppColors.darkOutline,
    success: Color(0xFF81C784),
    warning: Color(0xFFFFD54F),
    danger: Color(0xFFEF9A9A),
    info: Color(0xFF90CAF9),
    feeAccent: Color(0xFFCE93D8),
    scrim: AppColors.darkScrim,
  );

  /// Tenant branding hook: rebuild semantics with a school's primary color.
  AppSemantics withBrand(Color brandPrimary) => AppSemantics(
        primary: brandPrimary,
        onPrimary: onPrimary,
        surface: surface,
        surfaceVariant: surfaceVariant,
        background: background,
        text: text,
        textMuted: textMuted,
        border: border,
        success: success,
        warning: warning,
        danger: danger,
        info: info,
        feeAccent: feeAccent,
        scrim: scrim,
      );

  @override
  AppSemantics copyWith({
    Color? primary,
    Color? onPrimary,
    Color? surface,
    Color? surfaceVariant,
    Color? background,
    Color? text,
    Color? textMuted,
    Color? border,
    Color? success,
    Color? warning,
    Color? danger,
    Color? info,
    Color? feeAccent,
    Color? scrim,
  }) {
    return AppSemantics(
      primary: primary ?? this.primary,
      onPrimary: onPrimary ?? this.onPrimary,
      surface: surface ?? this.surface,
      surfaceVariant: surfaceVariant ?? this.surfaceVariant,
      background: background ?? this.background,
      text: text ?? this.text,
      textMuted: textMuted ?? this.textMuted,
      border: border ?? this.border,
      success: success ?? this.success,
      warning: warning ?? this.warning,
      danger: danger ?? this.danger,
      info: info ?? this.info,
      feeAccent: feeAccent ?? this.feeAccent,
      scrim: scrim ?? this.scrim,
    );
  }

  @override
  AppSemantics lerp(AppSemantics? other, double t) {
    if (other == null) return this;
    return AppSemantics(
      primary: Color.lerp(primary, other.primary, t)!,
      onPrimary: Color.lerp(onPrimary, other.onPrimary, t)!,
      surface: Color.lerp(surface, other.surface, t)!,
      surfaceVariant: Color.lerp(surfaceVariant, other.surfaceVariant, t)!,
      background: Color.lerp(background, other.background, t)!,
      text: Color.lerp(text, other.text, t)!,
      textMuted: Color.lerp(textMuted, other.textMuted, t)!,
      border: Color.lerp(border, other.border, t)!,
      success: Color.lerp(success, other.success, t)!,
      warning: Color.lerp(warning, other.warning, t)!,
      danger: Color.lerp(danger, other.danger, t)!,
      info: Color.lerp(info, other.info, t)!,
      feeAccent: Color.lerp(feeAccent, other.feeAccent, t)!,
      scrim: Color.lerp(scrim, other.scrim, t)!,
    );
  }
}
