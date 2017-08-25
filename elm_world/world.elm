import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
import Array

--state
type alias Model =
  { pos_x : Int
  , pos_y : Int
  , world : Array.Array Int
  , abundance_at_xy : Int
  , available_steps : Int
  , error : String
  }

load_world_data: Cmd Msg
load_world_data = Http.send WorldDataArrived (Http.getString "/get_world")

init: (Model, Cmd Msg)
init = (Model 0 0 Array.empty 0 0 "", load_world_data)

--names of things that can happen
type Msg = MoveUp | MoveDown | MoveLeft | MoveRight | WorldDataArrived (Result Http.Error String)

nav_bar: Html Msg
nav_bar =
  nav [class "navbar navbar-inverse"]
  [ div [class "container-fluid"]
    [ div [class "navbar-header"]
      [ a [class "navbar-brand", href "/home"] [text "Browser Game"]
      ]
    , ul [class "nav navbar-nav"]
      [ li [] [a [href "/home"] [text "Home"]]
      , li [class "active"] [a [href "/world"] [text "World"]]
      ]
    , ul [class "nav navbar-nav navbar-right"]
      [ li [] [a [href "/unlogin"] [span [class "glyphicon glyphicon-log-out"] [], text " Unlogin"]]
      ]
    ]
  ]

type BorderType = NoBorder | NormalBorder | CurrentBorder

get_tile: BorderType -> Maybe Int -> Html Msg
get_tile border_style mi= 
  let
    the_style =
      if border_style==NoBorder then
        style [("margin", "6px")]
      else if border_style==NormalBorder then
        style [("margin", "2px"), ("border-style", "solid"), ("border-width", "4px")]
      else if border_style==CurrentBorder then
        style [("margin", "2px"), ("border-style", "solid"), ("border-width", "4px"), ("border-color", "blue")]
      else
        style []
  in
    case mi of
      Nothing -> img [src "/files/empty32.png", the_style] []
      Just 0 -> img [src "/files/no_resource.png", the_style] []
      Just 1 -> img [src "/files/resource_a.png", the_style] []
      Just 2 -> img [src "/files/resource_b.png", the_style] []
      Just 3 -> img [src "/files/resource_c.png", the_style] []
      Just _ -> img [src "/files/empty32.png", the_style] []

world_map: Model -> Html Msg
world_map model =
  div []
  [ div [style [("float", "center")]]
    [ get_tile NoBorder Nothing
    , model.world |> Array.get 0 |> get_tile NormalBorder
    , model.world |> Array.get 1 |> get_tile NormalBorder
    , model.world |> Array.get 2 |> get_tile NormalBorder
    , get_tile NoBorder Nothing
    ]
  , div [style [("float", "center")]]
    [ model.world |> Array.get 3 |> get_tile NormalBorder
    , model.world |> Array.get 4 |> get_tile NormalBorder
    , model.world |> Array.get 5 |> get_tile NormalBorder
    , model.world |> Array.get 6 |> get_tile NormalBorder
    , model.world |> Array.get 7 |> get_tile NormalBorder]
  , div [style [("float", "center")]]
    [ model.world |> Array.get 8 |> get_tile NormalBorder
    , model.world |> Array.get 9 |> get_tile NormalBorder
    , model.world |> Array.get 10 |> get_tile CurrentBorder
    , model.world |> Array.get 11 |> get_tile NormalBorder
    , model.world |> Array.get 12 |> get_tile NormalBorder]
  , div [style [("float", "center")]]
    [ model.world |> Array.get 13 |> get_tile NormalBorder
    , model.world |> Array.get 14 |> get_tile NormalBorder
    , model.world |> Array.get 15 |> get_tile NormalBorder
    , model.world |> Array.get 16 |> get_tile NormalBorder
    , model.world |> Array.get 17 |> get_tile NormalBorder]
  , div [style [("float", "center")]]
    [ get_tile NoBorder Nothing
    , model.world |> Array.get 18 |> get_tile NormalBorder
    , model.world |> Array.get 19 |> get_tile NormalBorder
    , model.world |> Array.get 20 |> get_tile NormalBorder
    , get_tile NoBorder Nothing
    ]
  ]

navigator: Model -> Html Msg
navigator model =
  div []
  [ div [style [("float", "center"), ("margin", "8px")]]
    [ button [class "btn btn-default", onClick MoveUp] [img [src "/files/arrow_up.png"] []]
    , button [class "btn btn-default", onClick MoveDown] [img [src "/files/arrow_down.png"] []]
    , button [class "btn btn-default", onClick MoveLeft] [img [src "/files/arrow_left.png"] []]
    , button [class "btn btn-default", onClick MoveRight] [img [src "/files/arrow_right.png"] []]
    ]
  ]

--how it looks
view: Model -> Html Msg
view model =
  div []
  [ node "link" [ rel "stylesheet", href "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css"] []
  , nav_bar
  , div [class "container", style [("background-color", "#EFFFEF"), ("border-radius", "6px")]]
    [ world_map model
    , navigator model
    ]
  ]


--load_world_data: Cmd Msg
--load_world_data = Http.send WorldDataArrived (Http.getString "/get_world")

update: Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    MoveUp -> (model, Cmd.none)
    MoveDown -> (model, Cmd.none)
    MoveLeft -> (model, Cmd.none)
    MoveRight -> (model, Cmd.none)
    WorldDataArrived (Ok json_string) -> 
      let
        pos_x = case Decode.decodeString (Decode.field "Pos_x" Decode.int) json_string of
          Ok(x) -> x
          Err(_) -> -1
        pos_y = case Decode.decodeString (Decode.field "Pos_y" Decode.int) json_string of
          Ok(x) -> x
          Err(_) -> -1
        world = case Decode.decodeString (Decode.field "World_array" (Decode.array Decode.int)) json_string of
          Ok(x) -> x
          Err(_) -> Array.empty
        abundance_at_xy = case Decode.decodeString (Decode.field "Abundance_at_xy" Decode.int) json_string of
          Ok(x) -> x
          Err(_) -> -1
        available_steps = case Decode.decodeString (Decode.field "Available_steps" Decode.int) json_string of
          Ok(x) -> x
          Err(_) -> -1
      in
        ({model | pos_x=pos_x, pos_y=pos_y, world=world, abundance_at_xy=abundance_at_xy, available_steps=available_steps}, Cmd.none)
    WorldDataArrived (Err err) -> case err of 
      Http.BadUrl s -> ({model | error="BadUrl: "++s}, Cmd.none)
      Http.Timeout -> ({model | error="Timeout"}, Cmd.none)
      Http.NetworkError -> ({model | error="NetworkError"}, Cmd.none)
      Http.BadStatus _ -> ({model | error="BadStatus"}, Cmd.none)
      Http.BadPayload _ _ -> ({model | error="BadPayload"}, Cmd.none)
          

--events to be notified of
subscriptions: Model -> Sub Msg
subscriptions model=
  Sub.none

main: Program Never Model Msg
main =
  program
    {init=init
    ,view=view
    ,update=update
    ,subscriptions=subscriptions
    }
