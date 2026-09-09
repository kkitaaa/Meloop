/// Error controlado producido al comunicarse con el API Gateway.
class ApiException implements Exception {
  const ApiException({
    required this.message,
    this.statusCode,
    this.code,
    this.details,
  });

  final String message;
  final int? statusCode;
  final String? code;
  final Object? details;

  @override
  String toString() {
    final status = statusCode == null ? '' : ' (HTTP $statusCode)';
    final errorCode = code == null ? '' : ' [$code]';
    return 'ApiException$errorCode$status: $message';
  }
}
