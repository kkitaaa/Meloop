class ApiResponse<T> {
  final bool success;
  final T? data;
  final ApiError? error;

  const ApiResponse({required this.success, this.data, this.error});

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Object? json) parseData,
  ) {
    final success = json['success'];
    if (success is! bool) {
      throw const FormatException('La respuesta no incluye success válido.');
    }

    final data = json['data'];
    final errorJson = json['error'];
    if (errorJson != null && errorJson is! Map<String, dynamic>) {
      throw const FormatException('La respuesta incluye un error inválido.');
    }

    return ApiResponse<T>(
      success: success,
      data: success && data != null ? parseData(data) : null,
      error: errorJson == null
          ? null
          : ApiError.fromJson(errorJson as Map<String, dynamic>),
    );
  }
}

class ApiError {
  final String code;
  final String message;

  const ApiError({required this.code, required this.message});

  factory ApiError.fromJson(Map<String, dynamic> json) {
    final code = json['code'];
    final message = json['message'];
    if (code is! String || message is! String) {
      throw const FormatException('La respuesta incluye un error inválido.');
    }

    return ApiError(code: code, message: message);
  }
}
