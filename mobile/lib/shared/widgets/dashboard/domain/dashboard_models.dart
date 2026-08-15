/// Models the `/dashboard/summary` payload.
class DashboardSummary {
  const DashboardSummary({
    required this.students,
    required this.teachers,
    this.attendanceTodayPercent,
    this.pendingFeesCount = 0,
    this.pendingFeesAmountInr = 0,
    this.revenueCollectedInr = 0,
    this.collectionRate = 0,
  });

  final int students;
  final int teachers;
  final double? attendanceTodayPercent;
  final int pendingFeesCount;
  final int pendingFeesAmountInr;
  final int revenueCollectedInr;
  final double collectionRate;

  factory DashboardSummary.fromJson(Map<String, dynamic> json) {
    final pending = json['pending_fees'] as Map<String, dynamic>? ?? const {};
    final revenue = json['revenue'] as Map<String, dynamic>? ?? const {};
    return DashboardSummary(
      students: json['students'] as int? ?? 0,
      teachers: json['teachers'] as int? ?? 0,
      attendanceTodayPercent: (json['attendance_today_percent'] as num?)?.toDouble(),
      pendingFeesCount: pending['count'] as int? ?? 0,
      pendingFeesAmountInr: pending['amount_inr'] as int? ?? 0,
      revenueCollectedInr: revenue['collected_inr'] as int? ?? 0,
      collectionRate: (revenue['collection_rate'] as num?)?.toDouble() ?? 0,
    );
  }
}
