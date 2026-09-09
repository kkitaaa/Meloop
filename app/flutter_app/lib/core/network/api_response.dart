/// Envoltura de respuesta común definida por el API Gateway.
class ApiResponse<T> {
  const ApiResponse({required this.success, this.data, this.error});

  final bool success;
  final T? data;
  final ApiError? error;

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Object? json) fromData,
  ) {
    final rawError = json['error'];
    return ApiResponse<T>(
      success: json['success'] as bool? ?? false,
      data: json['data'] == null ? null : fromData(json['data']),
      error: rawError is Map<String, dynamic>
          ? ApiError.fromJson(rawError)
          : null,
    );
  }
}

class ApiError {
  const ApiError({required this.code, required this.message, this.details});

  final String code;
  final String message;
  final Object? details;

  factory ApiError.fromJson(Map<String, dynamic> json) => ApiError(
        code: json['code'] as String? ?? 'UNKNOWN_ERROR',
        message: json['message'] as String? ?? 'Ocurrió un error inesperado.',
        details: json['details'],
      );
}
