'use client';

import { useEffect, useState } from 'react';
import { Container, Title, Paper, Text, Group, Button, Stack, Grid } from '@mantine/core';

export function BattleWidget() {
  const [initialized, setInitialized] = useState(false);
  const [ffish, setFfish] = useState<any>(null);
  const [board, setBoard] = useState<any>(null);
  const [fen, setFen] = useState('');

  useEffect(() => {
    // Fairy-Stockfishの初期化
    const loadFfish = async () => {
      const ffishModule = await import('ffish');
      const ffish = await ffishModule.default();
      
      setFfish(ffish);
      
      // 9x9のチェスVS将棋のカスタムボードを作成
      // shogiバリアントをベースに使用（9x9ボード）
      const newBoard = new ffish.Board('shogi');
      
      // カスタムFEN: チェスの駒（大文字）vs 将棋の駒（小文字）
      // 9x9ボードでチェス側（上段）と将棋側（下段）
      // チェス側: RNBQKQBNR + ポーン
      // 将棋側: 通常の将棋の配置
      const customFen = 'lnsgkgsnl/1r5b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/RNBQKQBNR[-] w - 1';
      
      try {
        newBoard.setFen(customFen);
      } catch (e) {
        // カスタムFENが無効な場合は通常の将棋盤面を使用
        console.warn('カスタムFEN設定失敗、デフォルト盤面を使用');
      }
      
      setBoard(newBoard);
      setFen(newBoard.fen());
      setInitialized(true);
    };
    
    loadFfish();

    return () => {
      // クリーンアップ
      if (board) {
        board.delete();
      }
    };
  }, []);

  const getLegalMoves = () => {
    if (board) {
      try {
        return board.legalMoves().split(' ').filter((move: string) => move);
      } catch (e) {
        return [];
      }
    }
    return [];
  };

  const makeMove = (move: string) => {
    if (board) {
      try {
        if (board.isLegal(move)) {
          board.push(move);
          setFen(board.fen());
        }
      } catch (e) {
        console.error('着手エラー:', e);
      }
    }
  };

  const resetBoard = () => {
    if (!ffish || !board) return;
    
    const customFen = 'lnsgkgsnl/1r5b1/ppppppppp/9/9/9/PPPPPPPPP/1B5R1/RNBQKQBNR[-] w - 1';
    try {
      board.setFen(customFen);
      setFen(board.fen());
    } catch (e) {
      console.error('リセットエラー:', e);
    }
  };

  const renderBoard = () => {
    if (!board) return null;

    try {
      const boardStr = board.toString();
      const rows = boardStr.trim().split('\n');
      
      return (
        <div style={{ display: 'inline-block' }}>
          {rows.map((row: string, rowIndex: number) => (
            <div key={rowIndex} style={{ display: 'flex' }}>
              {row.split('').filter((c: string) => c !== ' ').map((piece: string, colIndex: number) => {
                const isLight = (rowIndex + colIndex) % 2 === 0;
                return (
                  <div
                    key={colIndex}
                    style={{
                      width: '50px',
                      height: '50px',
                      backgroundColor: isLight ? '#f0d9b5' : '#b58863',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '24px',
                      fontWeight: 'bold',
                      border: '1px solid #999',
                      userSelect: 'none',
                    }}
                  >
                    {piece === '.' ? '' : piece}
                  </div>
                );
              })}
            </div>
          ))}
        </div>
      );
    } catch (e) {
      return <Text c="red">盤面表示エラー</Text>;
    }
  };

  if (!initialized) {
    return (
      <Container size="lg" my={40}>
        <Paper shadow="md" p="xl" withBorder>
          <Text>Fairy-Stockfishを初期化中...</Text>
        </Paper>
      </Container>
    );
  }

  const legalMoves = getLegalMoves();

  return (
    <Container size="xl" my={40}>
      <Stack gap="xl">
        <div>
          <Title order={1} mb="md">
            チェス VS 将棋
          </Title>
          <Text c="dimmed">9×9ボード - チェス（白・大文字）vs 将棋（黒・小文字）</Text>
          <Text size="sm" c="dimmed" mt="xs">
            チェス側: キングの両脇にクイーン、通常のチェス駒配置 | 将棋側: 通常の将棋駒配置
          </Text>
        </div>

        <Paper shadow="md" p="xl" withBorder>
          <Stack gap="md">
            <Group justify="center">
              {renderBoard()}
            </Group>

            <Group justify="center">
              <Button variant="filled" onClick={resetBoard}>
                盤面リセット
              </Button>
            </Group>

            <div>
              <Text fw={500} mb="xs">
                現在の盤面（FEN表記）:
              </Text>
              <Paper p="sm" bg="gray.0">
                <Text size="sm" style={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>
                  {fen}
                </Text>
              </Paper>
            </div>

            <div>
              <Text fw={500} mb="xs">
                合法手（最初の15個）:
              </Text>
              <Paper p="sm" bg="gray.0">
                {legalMoves.length > 0 ? (
                  <Group gap="xs">
                    {legalMoves.slice(0, 15).map((move: string, index: number) => (
                      <Button
                        key={index}
                        size="xs"
                        variant="light"
                        onClick={() => makeMove(move)}
                      >
                        {move}
                      </Button>
                    ))}
                    {legalMoves.length > 15 && (
                      <Text size="sm" c="dimmed">
                        ... 他 {legalMoves.length - 15} 手
                      </Text>
                    )}
                  </Group>
                ) : (
                  <Text size="sm" c="dimmed">
                    合法手がありません（ゲーム終了）
                  </Text>
                )}
              </Paper>
            </div>

            <div>
              <Text size="sm" c="dimmed">
                ※ 合法手のボタンをクリックして駒を動かすことができます
              </Text>
              <Text size="sm" c="dimmed">
                ※ 大文字 = チェスの駒（白）、小文字 = 将棋の駒（黒）
              </Text>
            </div>
          </Stack>
        </Paper>
      </Stack>
    </Container>
  );
}
