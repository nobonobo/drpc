# 分散RPCツールキット

drpc: distributed rpc toolkit for golang

## 用語

- ノード(Node) クラスタを構成する１プロセス
- プロバイド一覧(Provides) 提供するサービス群
- サービスマップ(ServiceMap) クラスタ内のどのノードでどのサービス稼働中かを示す情報セット
- マスター クラスタ内のリーダーノード、２つ存在することは禁止
- スレーブ マスター以外のノード

## サービスインターフェース

- NodeService: １ノードに１つだけ必ずサポートする 
- NamingService: サービスのアドレスや稼動状態を管理する
- 追加予定

### NodeService: (すべてのノードがサポートする)
- Invite: NamingServiceへの接続招待
    - 既存のNamingServiceへの接続を破棄して呼ばれたノードは指定先に接続する
- Bye: 指定ノードとの接続切り離し（動作は継続）

### NamingService: 名前引きサービス
- Register: 登録
- Query: サービス名でサポートノード一覧を得る

