import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
--import Json.Encode as Encode
import Array

type alias PositionData =
  { pos_x : Int
  , pos_y : Int
  , world : Array.Array Int
  , abundance_at_xy : Int
  , available_steps : Int
  }

--state
type alias Model =
  { position_data : PositionData
  , error : String
  }

--(Decode.array Decode.int)
position_data_decoder : Decode.Decoder PositionData
position_data_decoder = 
  let
    map5 = Decode.map5
    field = Decode.field
    array = Decode.array
    int = Decode.int
  in
    map5 PositionData
      (field "X" int)
      (field "Y" int)
      (field "World_array" (array int))
      (field "Abundance_at_xy" int)
      (field "Available_steps" int)

load_world_data: Cmd Msg
load_world_data = Http.send WorldDataArrived (Http.get "/get_world" position_data_decoder)

type Direction = Up | Down | Left | Right

--get_post_body_for_move_request : Direction -> Http.Body
--get_post_body_for_move_request direction =
--  let
--    direction_string = case direction of
--      Up -> "Up"
--      Down -> "Down"
--      Left -> "Left"
--      Right -> "Right"
--  in
--    [("direction", Encode.string direction_string )]
--    |> Encode.object
--    |> Http.jsonBody

--get_post_body_for_move_request : Direction -> Http.Body
--get_post_body_for_move_request direction =
--  let
--    direction_string = case direction of
--      Up -> "Up"
--      Down -> "Down"
--      Left -> "Left"
--      Right -> "Right"
--  in
--    Http.multipartBody [Http.stringPart "direction" direction_string]

get_post_body_for_move_request : Direction -> Http.Body
get_post_body_for_move_request direction =
  let
    direction_string = case direction of
      Up -> "Up"
      Down -> "Down"
      Left -> "Left"
      Right -> "Right"
  in
    Http.stringBody "application/x-www-form-urlencoded" ("direction="++direction_string)

send_move_request: Direction -> Cmd Msg
send_move_request direction =
  Http.send WorldDataArrived (Http.post "/get_world" (get_post_body_for_move_request direction) position_data_decoder)

init: (Model, Cmd Msg)
init = (Model (PositionData 0 0 Array.empty 0 0) "", load_world_data)

--names of things that can happen
type Msg = Move (Direction) | WorldDataArrived (Result Http.Error PositionData)

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
    , model.position_data.world |> Array.get 0 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 1 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 2 |> get_tile NormalBorder
    , get_tile NoBorder Nothing
    ]
  , div [style [("float", "center")]]
    [ model.position_data.world |> Array.get 3 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 4 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 5 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 6 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 7 |> get_tile NormalBorder]
  , div [style [("float", "center")]]
    [ model.position_data.world |> Array.get 8 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 9 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 10 |> get_tile CurrentBorder
    , model.position_data.world |> Array.get 11 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 12 |> get_tile NormalBorder]
  , div [style [("float", "center")]]
    [ model.position_data.world |> Array.get 13 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 14 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 15 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 16 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 17 |> get_tile NormalBorder]
  , div [style [("float", "center")]]
    [ get_tile NoBorder Nothing
    , model.position_data.world |> Array.get 18 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 19 |> get_tile NormalBorder
    , model.position_data.world |> Array.get 20 |> get_tile NormalBorder
    , get_tile NoBorder Nothing
    ]
  ]

navigator: Model -> Html Msg
navigator model =
  div []
  [ div [style [("float", "center"), ("margin", "8px")]]
    [ button [class "btn btn-default", onClick (Move Up)] [img [src "/files/arrow_up.png"] []]
    , button [class "btn btn-default", onClick (Move Down)] [img [src "/files/arrow_down.png"] []]
    , button [class "btn btn-default", onClick (Move Left)] [img [src "/files/arrow_left.png"] []]
    , button [class "btn btn-default", onClick (Move Right)] [img [src "/files/arrow_right.png"] []]
    ]
  ]

error_message: Model -> Html Msg
error_message model =
  if model.error=="" then
    div [] []
  else
    div [class "alert alert-warning"] 
    [ strong [] [text ("Warning! ("++model.error++")")]
    , text " Data could not be read, please consider"
    , a [href "/login", class "alert-link"] [text " logging in"]
    , text " again."
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
    , error_message model
    ]
  ]

--load_world_data: Cmd Msg
--load_world_data = Http.send WorldDataArrived (Http.getString "/get_world")

update: Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    Move direction -> (model, send_move_request direction)
    --MoveUp -> (model, Cmd.none)
    --MoveDown -> (model, Cmd.none)
    --MoveLeft -> (model, Cmd.none)
    --MoveRight -> (model, Cmd.none)
    WorldDataArrived (Ok position_data) -> ({model | position_data=position_data}, Cmd.none)
    WorldDataArrived (Err err) -> case err of 
      Http.BadUrl s -> ({model | error="BadUrl: "++s}, Cmd.none)
      Http.Timeout -> ({model | error="Timeout"}, Cmd.none)
      Http.NetworkError -> ({model | error="NetworkError"}, Cmd.none)
      Http.BadStatus _ -> ({model | error="BadStatus"}, Cmd.none)
      Http.BadPayload s _ -> ({model | error="BadPayload: "++s}, Cmd.none)
          

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
