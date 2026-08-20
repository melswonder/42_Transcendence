// アプリケーション固有の「ユーザーが達成したい目的（シナリオ）」を定義し、進行する層です。
// ドメイン層のルールを呼び出しながら、データの流れをオーケストレーション（指揮）します。
// 【概念の例：書籍の購入】
// シナリオ: 「カートの中身を購入する」
// 進行手順:
//
//	カートのデータを取得する
//	ドメイン層のルールを使って合計金額（割引など）を計算する
//	決済処理を依頼する
//	注文履歴を保存し、カートを空にする
//
// 「何をすべきか」という手順は知っていますが、「データがどのDB（MySQLかOracleか）に保存されるか」は知りません。
package usecase

// usecase層の受け口
type Dependencies struct {
	PingRepository PingRepository
	AuthRepository AuthRepository
	GoogleOAuth    OAuthProvider
}

type Services struct {
	Ping *PingUsecase
	Auth *AuthUsecase
}

func NewServices(dependencies Dependencies) Services {
	return Services{
		Ping: NewPingUsecase(dependencies.PingRepository),
		Auth: NewAuthUsecase(dependencies.GoogleOAuth, dependencies.AuthRepository),
	}
}
