/// Implementación temporal para la demostración visual de interacciones.
///
/// Conserva la interfaz que usará el cliente REST cuando el API Gateway de
/// publicaciones esté disponible, pero por ahora no realiza llamadas de red.
class PostInteractionApi {
  const PostInteractionApi();

  Future<void> setLike({required String postId, required bool liked}) async {
    await Future<void>.delayed(const Duration(milliseconds: 180));
  }

  Future<void> setBookmark({
    required String postId,
    required bool saved,
  }) async {
    await Future<void>.delayed(const Duration(milliseconds: 180));
  }
}
