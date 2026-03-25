# Munchkin Game WebSocket Protocol Specification

## Overview
This document defines the WebSocket message protocol for real-time lobby management and gameplay communication between the Munchkin game server and clients. The WebSocket endpoint is `/lobby/{id}/ws` (authenticated with JWT).

## Connection Lifecycle
1. **Authentication**: Client authenticates via HTTP `POST /auth/login` to get JWT
2. **Connection**: Client connects to `ws://{host}:1337/lobby/{id}/ws?token={jwt_token}` with JWT in query parameter
3. **Lobby Join**: Upon connection, server sends `LOBBY_STATE` with current lobby status
4. **Lobby Management**: Players can set ready status, chat, and host can change settings/kick players
5. **Game Start**: Host or automatic trigger starts game → server sends `GAME_STARTING` countdown → `GAME_STARTED` with initial state
6. **Gameplay**: Real-time messages flow via WebSocket
7. **Disconnection**: Client reconnects using same JWT, receives `LOBBY_STATE` (if in lobby) or `GAME_STATE` (if in game)

## Message Format
All messages are JSON objects with a `type` field and `data` payload:
```json
{
  "type": "MESSAGE_TYPE",
  "data": { ... }
}
```

## Client → Server Messages

### Game Actions
```json
// Player takes an action during their turn
{
  "type": "PLAYER_ACTION",
  "data": {
    "action": "OPEN_DOOR" | "LOOK_FOR_TROUBLE" | "LOOT_ROOM" | "END_TURN",
    "timestamp": "2024-03-13T20:30:00Z"
  }
}

// Play a card from hand
{
  "type": "PLAY_CARD",
  "data": {
    "card_id": "monster_goblin_001",
    "target_player_id": "player-uuid",  // optional, for curses
    "additional_data": {}  // card-specific data
  }
}

// Response to combat interaction
{
  "type": "COMBAT_RESPONSE",
  "data": {
    "response": "ACCEPT_ALLIANCE" | "DECLINE_ALLIANCE" | "FLEE" | "PLAY_CARD",
    "card_id": "optional-card-id",
    "negotiation_terms": {}  // if accepting alliance
  }
}

// Offer/respond to negotiation
{
  "type": "NEGOTIATION",
  "data": {
    "action": "OFFER" | "ACCEPT" | "REJECT" | "COUNTER_OFFER",
    "negotiation_id": "uuid",
    "terms": {
      "treasure_split": {"player_id": 2, "other_player_id": 1},  // e.g., 2:1 split
      "card_offers": ["card_id_1", "card_id_2"]  // cards offered
    }
  }
}

// Use class ability
{
  "type": "USE_ABILITY",
  "data": {
    "ability": "THIEF_STEAL" | "WARRIOR_DISCARD" | "MAGE_CHARM" | "CLERIC_RESURRECT",
    "target_player_id": "player-uuid",  // for thief steal
    "card_ids": ["card1", "card2"]  // for warrior/mage discard
  }
}
```

## Lobby Events (Server → Client)

### Lobby State Updates
These messages are sent to all clients in the lobby while waiting for the game to start.

```json
// Lobby full state snapshot (when player joins or reconnects)
{
  "type": "LOBBY_STATE",
  "data": {
    "lobby_id": "lobby-uuid",
    "host_id": "player-uuid",
    "players": [
      {
        "id": "player-uuid",
        "name": "Player Name",
        "is_host": true,
        "is_ready": false,
        "avatar": "avatar_id"  // optional
      }
    ],
    "game_in_progress": false,
    "max_players": 6,
    "current_players": 3,
    "settings": {
      "timer_enabled": true,
      "turn_time_limit": 120,  // seconds
      "combat_interaction_time": 30  // seconds
    }
  }
}

// Player joined the lobby
{
  "type": "PLAYER_JOINED",
  "data": {
    "player_id": "player-uuid",
    "name": "Player Name",
    "is_host": false,
    "current_players": 4  // updated count
  }
}

// Player left the lobby
{
  "type": "PLAYER_LEFT",
  "data": {
    "player_id": "player-uuid",
    "name": "Player Name",
    "reason": "DISCONNECTED" | "KICKED" | "VOLUNTARY",
    "current_players": 3  // updated count
  }
}

// Player changed ready status
{
  "type": "PLAYER_READY_CHANGE",
  "data": {
    "player_id": "player-uuid",
    "is_ready": true
  }
}

// Host changed lobby settings
{
  "type": "LOBBY_SETTINGS_CHANGE",
  "data": {
    "settings": {
      "timer_enabled": false,
      "turn_time_limit": 90,
      "combat_interaction_time": 45
    },
    "changed_by": "player-uuid"
  }
}

// Game starting countdown
{
  "type": "GAME_STARTING",
  "data": {
    "countdown": 10,  // seconds until game starts
    "reason": "ALL_READY" | "HOST_FORCED" | "TIMER_EXPIRED"
  }
}

// Game started
{
  "type": "GAME_STARTED",
  "data": {
    "first_player_id": "player-uuid",
    "initial_hand_size": 8,
    "initial_state": {
      // Full game state as defined in GAME_STATE message
    }
  }
}
```

### Lobby Actions (Client → Server)
These messages are sent by clients while in the lobby.

```json
// Set ready status
{
  "type": "SET_READY",
  "data": {
    "is_ready": true
  }
}

// Change lobby settings (host only)
{
  "type": "CHANGE_SETTINGS",
  "data": {
    "timer_enabled": true,
    "turn_time_limit": 120,
    "combat_interaction_time": 30
  }
}

// Kick player (host only)
{
  "type": "KICK_PLAYER",
  "data": {
    "player_id": "player-uuid",
    "reason": "AFK" | "DISRUPTIVE" | "OTHER"
  }
}

// Start game (host only)
{
  "type": "START_GAME",
  "data": {
    "force_start": false  // start even if not all players ready
  }
}

// Send chat message
{
  "type": "LOBBY_CHAT",
  "data": {
    "message": "Hello everyone!",
    "timestamp": "2024-03-13T20:30:00Z"
  }
}

// Receive chat message (Server → Client)
{
  "type": "LOBBY_CHAT_MESSAGE",
  "data": {
    "player_id": "player-uuid",
    "player_name": "Player Name",
    "message": "Hello everyone!",
    "timestamp": "2024-03-13T20:30:00Z"
  }
}
```

## Server → Client Messages (Gameplay)
```json
// Full game state (on connection or major change)
{
  "type": "GAME_STATE",
  "data": {
    "game_id": "lobby-id",
    "players": [
      {
        "id": "player-uuid",
        "name": "Player Name",
        "level": 1,
        "race": "ELF" | "DWARF" | "HALFLING" | null,
        "class": "WARRIOR" | "THIEF" | "MAGE" | "CLERIC" | null,
        "sex": "MALE" | "FEMALE",
        "has_hybrid_race": false,
        "has_hybrid_class": false,
        "equipment": [
          {
            "card_id": "item_broad_sword_001",
            "slot": "HAND_1",
            "is_equipped": true
          }
        ],
        "hand": ["card_id_1", "card_id_2", ...],
        "is_dead": false
      }
    ],
    "current_turn": {
      "player_id": "player-uuid",
      "phase": "OPEN_DOOR" | "LOOK_FOR_TROUBLE" | "LOOT_ROOM" | "CHARITY",
      "phase_start_time": "2024-03-13T20:30:00Z",
      "time_remaining": 60  // seconds
    },
    "combat": {
      "active": true,
      "monsters": [
        {
          "card_id": "monster_goblin_001",
          "level": 1,
          "treasures": 1,
          "levels_gained": 1,
          "flee_penalty": "LOSE_ITEM",
          "flee_modifier": 0
        }
      ],
      "player_force": 5,
      "monster_force": 4,
      "ally": "player-uuid" | null,
      "interaction_window_open": true,
      "interaction_window_ends": "2024-03-13T20:31:00Z"
    } | null,
    "decks": {
      "dungeon_remaining": 85,
      "treasure_remaining": 65,
      "dungeon_discard": ["curse_lose_level_001"],
      "treasure_discard": []
    },
    "winner": "player-uuid" | null
  }
}

// Turn phase change notification
{
  "type": "TURN_PHASE_CHANGE",
  "data": {
    "player_id": "player-uuid",
    "phase": "OPEN_DOOR" | "LOOK_FOR_TROUBLE" | "LOOT_ROOM" | "CHARITY",
    "result": {
      "drawn_card": "monster_goblin_001" | null,
      "combat_triggered": true | false,
      "charity_cards": 2  // number of cards to give away
    }
  }
}

// Combat start
{
  "type": "COMBAT_START",
  "data": {
    "monster": {
      "card_id": "monster_goblin_001",
      "level": 1,
      "treasures": 1,
      "levels_gained": 1
    },
    "player_force": 5,
    "interaction_window_duration": 30  // seconds
  }
}

// Combat resolution
{
  "type": "COMBAT_RESOLUTION",
  "data": {
    "result": "VICTORY" | "DEFEAT",
    "player_force": 8,
    "monster_force": 5,
    "rewards": {
      "treasures": ["item_broad_sword_001", "action_potion_001"],
      "levels_gained": 1,
      "ally_levels_gained": 1  // if ally is elf
    } | null,
    "penalty": {
      "type": "LOSE_LEVEL" | "LOSE_ITEM" | "DEATH",
      "details": "Lose your headgear"
    } | null
  }
}

// Card play result
{
  "type": "CARD_PLAY_RESULT",
  "data": {
    "player_id": "player-uuid",
    "card_id": "monster_goblin_001",
    "success": true,
    "effect": "MONSTER_ADDED_TO_COMBAT",
    "validation_error": "INVALID_PHASE" | "INVALID_TARGET" | "NOT_IN_HAND" | null
  }
}

// Player state update (partial)
{
  "type": "PLAYER_UPDATE",
  "data": {
    "player_id": "player-uuid",
    "changes": {
      "level": 2,
      "hand": ["added_card_id"],
      "equipment": ["removed_item_id"]
    }
  }
}

// Error message
{
  "type": "ERROR",
  "data": {
    "code": "INVALID_ACTION" | "NOT_YOUR_TURN" | "INVALID_PHASE" | "NETWORK_ERROR",
    "message": "Human-readable error",
    "recoverable": true,
    "suggested_action": "RETRY" | "WAIT" | "DISCONNECT"
  }
}
```

## Data Types Reference

### Lobby Types
```typescript
interface LobbyPlayer {
  id: string;
  name: string;
  is_host: boolean;
  is_ready: boolean;
  avatar?: string;  // avatar identifier
}

interface LobbySettings {
  timer_enabled: boolean;
  turn_time_limit: number;  // seconds
  combat_interaction_time: number;  // seconds
}

interface LobbyState {
  lobby_id: string;
  host_id: string;
  players: LobbyPlayer[];
  game_in_progress: boolean;
  max_players: number;  // 3-6
  current_players: number;
  settings: LobbySettings;
}
```

### Card Types (from §4 of game rules)
```typescript
interface CardBase {
  id: string;
  name: string;
  description: string;
  deck_type: "DUNGEON" | "TREASURE";
}

interface MonsterCard extends CardBase {
  type: "MONSTER";
  level: number;
  bonus_against_race?: "ELF" | "DWARF" | "HALFLING";
  bonus_against_class?: "WARRIOR" | "THIEF" | "MAGE" | "CLERIC";
  bonus_value: number;
  flee_penalty: "LOSE_LEVEL" | "LOSE_ITEM" | "DEATH" | "CURSE";
  flee_modifier: number;
  treasures: number;
  levels_gained: number;
}

interface ItemCard extends CardBase {
  type: "ITEM";
  bonus: number;
  gold_value: number;  // multiple of 100
  size: "NORMAL" | "BIG";
  slot: "HEAD" | "ARMOR" | "FOOT" | "HAND_1" | "HAND_2" | "TWO_HANDS" | "NONE";
  race_restriction?: "ELF" | "DWARF" | "HALFLING";
  class_restriction?: "WARRIOR" | "THIEF" | "MAGE" | "CLERIC";
  sex_restriction?: "MALE" | "FEMALE";
}

interface CurseCard extends CardBase {
  type: "CURSE";
  effect: "LOSE_HEADGEAR" | "LOSE_LEVEL" | "CHANGE_SEX" | "LOSE_RACE";
}

interface RaceCard extends CardBase {
  type: "RACE";
  race: "ELF" | "DWARF" | "HALFLING";
}

interface ClassCard extends CardBase {
  type: "CLASS";
  class: "WARRIOR" | "THIEF" | "MAGE" | "CLERIC";
}

interface ActionCard extends CardBase {
  type: "ACTION";
  playable_when: "DURING_YOUR_TURN" | "DURING_COMBAT" | "ANYTIME" | "IN_RESPONSE";
}
```

### Player State (from §5 of game rules)
```typescript
interface PlayerState {
  id: string;
  name: string;
  level: number;  // 1-10
  race: RaceCard | null;
  race2: RaceCard | null;  // only if hybrid_race
  class: ClassCard | null;
  class2: ClassCard | null;  // only if hybrid_class
  sex: "MALE" | "FEMALE";
  has_hybrid_race: boolean;
  has_hybrid_class: boolean;
  equipment: EquipmentSlot[];
  carried_items: ItemCard[];  // items in play but not equipped
  hand: string[];  // card IDs
  is_dead: boolean;
}

interface EquipmentSlot {
  slot: "HEAD" | "ARMOR" | "FOOT" | "HAND_1" | "HAND_2" | "TWO_HANDS";
  card_id: string | null;
  is_big_item: boolean;
}
```

## Sequence Examples

### 0. Lobby Interaction Flow
```
Player 1 creates lobby:
  Player 2 joins:
    ← LOBBY_STATE {...players: [P1, P2]...}
    ← PLAYER_JOINED {player_id: P2, name: "Player 2"}
  
  Player 3 joins:
    ← LOBBY_STATE {...players: [P1, P2, P3]...}
    ← PLAYER_JOINED {player_id: P3, name: "Player 3"}
  
  Player 2 sets ready:
    → SET_READY {is_ready: true}
    ← PLAYER_READY_CHANGE {player_id: P2, is_ready: true}
  
  Player 3 sets ready:
    → SET_READY {is_ready: true}
    ← PLAYER_READY_CHANGE {player_id: P3, is_ready: true}
  
  Player 1 (host) starts game:
    → START_GAME {force_start: false}
    ← GAME_STARTING {countdown: 10, reason: "ALL_READY"}
    ... countdown ...
    ← GAME_STARTED {...initial_state...}
```

### 1. Complete Turn Cycle
```
Client 1 (Player 1 turn):
  → PLAYER_ACTION {action: "OPEN_DOOR"}
  ← TURN_PHASE_CHANGE {phase: "OPEN_DOOR", result: {drawn_card: "monster_goblin_001", combat_triggered: true}}
  ← COMBAT_START {...}
  → COMBAT_RESPONSE {response: "FLEE"}
  ← COMBAT_RESOLUTION {result: "VICTORY", rewards: {...}}
  → PLAYER_ACTION {action: "CHARITY"}
  ← TURN_PHASE_CHANGE {phase: "CHARITY", result: {charity_cards: 2}}
  → PLAY_CARD {card_id: "item_broad_sword_001"} // give away card
  → PLAYER_ACTION {action: "END_TURN"}
  ← GAME_STATE {...} // now Player 2's turn
```

### 2. Multiplayer Combat Interaction
```
Player 1 draws monster
  ← COMBAT_START {...interaction_window_open: true...}

Player 2 (inactive):
  → PLAY_CARD {card_id: "action_duck_of_doom_001"} // help monster
  ← CARD_PLAY_RESULT {success: true, effect: "MONSTER_BONUS_ADDED"}

Player 3 (inactive):
  → PLAY_CARD {card_id: "monster_potted_plant_001"} // wandering monster
  ← CARD_PLAY_RESULT {success: true, effect: "MONSTER_ADDED_TO_COMBAT"}

Player 4 (inactive):
  → NEGOTIATION {action: "OFFER", terms: {treasure_split: {player_id: 1, other_player_id: 2}}}
  ← NEGOTIATION {action: "OFFER_RECEIVED", from_player: "player-4-uuid"}

Player 1:
  → COMBAT_RESPONSE {response: "ACCEPT_ALLIANCE", negotiation_id: "..."}
  ← COMBAT_RESOLUTION {result: "VICTORY", ...}
```

## Error Handling

### Reconnection Flow
1. Client loses connection
2. Client reconnects to same WebSocket URL with JWT
3. Server sends full `GAME_STATE` message
4. Client resumes from last known state

### Invalid Actions
Server responds with `ERROR` message containing:
- `code`: Specific error type
- `message`: Human-readable description
- `recoverable`: Whether client can retry
- `suggested_action`: What client should do next

### Validation Rules (from game rules)
- Turn order enforcement (§7)
- Phase restrictions (§7.1)
- Card play timing (§4.4)
- Combat interaction windows (§8.2)
- Equipment slot rules (§9)
- Level 10 victory condition (§2)

## Implementation Notes

### Client-Side (Godot C#)
```csharp
// Message structure
public record WebSocketMessage {
    public string Type { get; set; }
    public JsonElement Data { get; set; }
}

// Client sends
public enum ClientMessageType {
    // Lobby messages
    SET_READY,
    CHANGE_SETTINGS,
    KICK_PLAYER,
    START_GAME,
    LOBBY_CHAT,
    
    // Game messages  
    PLAYER_ACTION,
    PLAY_CARD,
    COMBAT_RESPONSE,
    NEGOTIATION,
    USE_ABILITY
}

// Client receives  
public enum ServerMessageType {
    // Lobby messages
    LOBBY_STATE,
    PLAYER_JOINED,
    PLAYER_LEFT,
    PLAYER_READY_CHANGE,
    LOBBY_SETTINGS_CHANGE,
    GAME_STARTING,
    GAME_STARTED,
    LOBBY_CHAT_MESSAGE,
    
    // Game messages
    GAME_STATE,
    TURN_PHASE_CHANGE,
    COMBAT_START,
    COMBAT_RESOLUTION,
    CARD_PLAY_RESULT,
    PLAYER_UPDATE,
    ERROR
}
```

### Server-Side (Golang)
```go
type WebSocketMessage struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}

// Server should:
// 1. Validate JWT on connection
// 2. Maintain game state per lobby
// 3. Broadcast state changes to all connected clients
// 4. Enforce game rules strictly
// 5. Handle reconnections gracefully
```

## Next Steps

### For Server Team:
1. **Extend WebSocket handler** at `/lobby/{id}/ws` to include lobby management
2. **Define Go structs** for new lobby message types (`LOBBY_STATE`, `PLAYER_JOINED`, etc.)
3. **Implement lobby state management** (player list, ready status, settings)
4. **Implement game start transition** from lobby to game state
5. **Create initial card database** with 10 sample cards

### For Client Team:
1. **Extend `MessageProtocol.cs`** to include new lobby message types
2. **Update `WebSocketClient.cs`** to handle lobby events and game start transition
3. **Create `LobbyManager.cs`** to manage lobby UI and state
4. **Build lobby UI** showing player list, ready status, chat, and game start controls
5. **Handle game start transition** from lobby UI to game board

## Version History
- **v1.1** (2025-03-25): Added lobby events and management messages
- **v1.0** (2024-03-13): Initial protocol specification based on Munchkin game rules document v2.0

## References
- Game Rules Document: `AGENTS.md` (Munchkin rules v2.0)
- Server API: Swagger docs at `http://90.28.104.14:1337/swagger/doc.json`
- Card System: Godot `card_3d` plugin in `/addons/card_3d/`