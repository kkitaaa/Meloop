class UserModel {
  const UserModel({
    required this.id,
    required this.username,
    required this.email,
  });

  final int id;
  final String username;
  final String email;

  factory UserModel.fromJson(Object? json) {
    final map = json as Map<String, dynamic>;
    return UserModel(
      id: map['id'] as int,
      username: map['username'] as String,
      email: map['email'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'username': username,
        'email': email,
      };
}
