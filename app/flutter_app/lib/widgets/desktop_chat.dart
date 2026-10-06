import 'package:flutter/material.dart';

/// Chat flotante de demostración para pantallas de escritorio.
/// No abre WebSockets todavía: los mensajes viven solo durante esta sesión.
class DesktopChat extends StatefulWidget {
  const DesktopChat({super.key});

  @override
  State<DesktopChat> createState() => _DesktopChatState();
}

class _DesktopChatState extends State<DesktopChat> {
  static const _headerColor = Color(0xFF1E7E70);
  static const _accentColor = Color(0xFF00AF92);

  final _messageController = TextEditingController();
  bool _isOpen = false;
  _Conversation? _conversation;

  final _conversations = <_Conversation>[
    _Conversation(
      name: 'marcelo.2306',
      preview: 'Oe podi darle like a lo q postie ono',
      time: 'Ahora',
      online: true,
      avatarColor: const Color(0xFF69493D),
      messages: [
        _ChatMessage('Oe podi darle like a lo q postie ono', false, '13:25'),
        _ChatMessage('XDDDD nadie te pesca', true, 'Ahora', status: 'Leído'),
        _ChatMessage('como so oe', false, 'Ahora'),
      ],
    ),
    _Conversation(
      name: 'Maxwell.gato',
      preview: 'Amigo me puedes mandar el link del gi...',
      time: '13:20',
      avatarColor: const Color(0xFF303030),
      messages: [_ChatMessage('Amigo me puedes mandar el link del gif?', false, '13:20')],
    ),
    _Conversation(
      name: 'Rigby.gato',
      preview: 'Enviaste una canción',
      time: 'Ayer',
      avatarColor: const Color(0xFFC08366),
      messages: [_ChatMessage('Te envié una canción', true, 'Ayer', status: 'Leído')],
    ),
    _Conversation(
      name: 'Usuario Bloqueado',
      preview: '',
      time: '',
      avatarColor: Colors.blueGrey,
      messages: [],
    ),
  ];

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  void _openConversation(_Conversation conversation) {
    setState(() => _conversation = conversation);
  }

  void _sendMessage() {
    final text = _messageController.text.trim();
    if (text.isEmpty || _conversation == null) return;
    setState(() {
      _conversation!.messages.add(_ChatMessage(text, true, 'Ahora', status: 'Enviado'));
      _conversation!.preview = text;
      _conversation!.time = 'Ahora';
    });
    _messageController.clear();
  }

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      elevation: 14,
      borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOut,
        width: 350,
        height: _isOpen ? 510 : 46,
        decoration: BoxDecoration(
          color: const Color(0xFFFAFAFA),
          borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
          boxShadow: [
            BoxShadow(color: Colors.black.withValues(alpha: .22), blurRadius: 18),
          ],
        ),
        child: Column(
          children: [
            InkWell(
              onTap: () => setState(() => _isOpen = !_isOpen),
              borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
              child: Container(
                height: 46,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: const BoxDecoration(
                  color: _headerColor,
                  borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
                ),
                child: Row(
                  children: [
                    if (_conversation != null)
                      IconButton(
                        tooltip: 'Volver a conversaciones',
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                        icon: const Icon(Icons.arrow_back, color: Colors.white, size: 21),
                        onPressed: () => setState(() => _conversation = null),
                      ),
                    if (_conversation != null) const SizedBox(width: 10),
                    const Icon(Icons.chat_bubble_outline, color: Colors.white, size: 19),
                    const SizedBox(width: 8),
                    const Text('Chat', style: TextStyle(color: Colors.white, fontSize: 17, fontWeight: FontWeight.w600)),
                    const Spacer(),
                    if (_conversation == null && _isOpen)
                      const CircleAvatar(radius: 5, backgroundColor: Color(0xFF36C1A4)),
                    Icon(_isOpen ? Icons.keyboard_arrow_down : Icons.keyboard_arrow_up, color: Colors.white),
                  ],
                ),
              ),
            ),
            if (_isOpen) Expanded(child: _conversation == null ? _inbox() : _chat()),
          ],
        ),
      ),
    );
  }

  Widget _inbox() {
    return ListView.separated(
      padding: const EdgeInsets.all(10),
      itemCount: _conversations.length,
      separatorBuilder: (_, _) => const SizedBox(height: 8),
      itemBuilder: (context, index) {
        final conversation = _conversations[index];
        return Material(
          color: Colors.white,
          child: InkWell(
            onTap: () => _openConversation(conversation),
            child: Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: .06), blurRadius: 5)],
              ),
              child: Row(
                children: [
                  _avatar(conversation),
                  const SizedBox(width: 11),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(conversation.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15)),
                        const SizedBox(height: 5),
                        Text(conversation.preview, overflow: TextOverflow.ellipsis, style: const TextStyle(color: Colors.black45, fontSize: 12)),
                      ],
                    ),
                  ),
                  Text(conversation.time, style: const TextStyle(color: Colors.black38, fontSize: 11)),
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _chat() {
    final conversation = _conversation!;
    return Column(
      children: [
        Container(
          color: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 9),
          child: Row(
            children: [
              _avatar(conversation, small: true),
              const SizedBox(width: 9),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(conversation.name, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                  if (conversation.online) const Text('En línea', style: TextStyle(color: _accentColor, fontSize: 11)),
                ],
              ),
            ],
          ),
        ),
        const Divider(height: 1),
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.all(12),
            itemCount: conversation.messages.length,
            itemBuilder: (context, index) => _bubble(conversation.messages[index]),
          ),
        ),
        Container(
          color: _accentColor,
          padding: const EdgeInsets.fromLTRB(10, 8, 6, 8),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _messageController,
                  onSubmitted: (_) => _sendMessage(),
                  decoration: const InputDecoration(
                    hintText: 'Escribe aquí...',
                    filled: true,
                    fillColor: Colors.white,
                    isDense: true,
                    contentPadding: EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                    border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(22)), borderSide: BorderSide.none),
                  ),
                ),
              ),
              IconButton(icon: const Icon(Icons.send_rounded, color: Colors.white), onPressed: _sendMessage),
            ],
          ),
        ),
      ],
    );
  }

  Widget _bubble(_ChatMessage message) {
    final color = message.mine ? _accentColor : _headerColor;
    return Align(
      alignment: message.mine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.fromLTRB(12, 9, 10, 6),
        constraints: const BoxConstraints(maxWidth: 255),
        decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(10)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(message.text, style: const TextStyle(color: Colors.white, fontSize: 13)),
            const SizedBox(height: 3),
            Text('${message.time}${message.status == null ? '' : ' · ${message.status}'}', style: TextStyle(color: Colors.white.withValues(alpha: .6), fontSize: 9)),
          ],
        ),
      ),
    );
  }

  Widget _avatar(_Conversation conversation, {bool small = false}) {
    final size = small ? 28.0 : 42.0;
    return Container(
      width: size,
      height: size,
      color: conversation.avatarColor,
      alignment: Alignment.center,
      child: Icon(Icons.person, size: small ? 17 : 25, color: Colors.white70),
    );
  }
}

class _Conversation {
  _Conversation({required this.name, required this.preview, required this.time, required this.avatarColor, required this.messages, this.online = false});

  final String name;
  String preview;
  String time;
  final Color avatarColor;
  final bool online;
  final List<_ChatMessage> messages;
}

class _ChatMessage {
  const _ChatMessage(this.text, this.mine, this.time, {this.status});

  final String text;
  final bool mine;
  final String time;
  final String? status;
}
