import 'package:flutter_app/core/network/api_client.dart';
import 'package:flutter_app/features/users/models/user_model.dart';

/// Operaciones del dominio de usuarios disponibles mediante el API Gateway.
class UserService {
  UserService(this._apiClient);

  final ApiClient _apiClient;

  Future<List<UserModel>> getUsers() async {
    final response = await _apiClient.get<List<UserModel>>(
      '/users',
      fromData: (json) => (json as List<Object?>)
          .map(UserModel.fromJson)
          .toList(growable: false),
    );
    return response.data ?? const [];
  }

  Future<UserModel> createUser({
    required String username,
    required String email,
  }) async {
    final response = await _apiClient.post<UserModel>(
      '/users',
      body: {'username': username, 'email': email},
      fromData: UserModel.fromJson,
    );
    return response.data!;
  }
}
