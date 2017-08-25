import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode

type alias Player = 
  { name : String
  , pos_x : Int
  , pos_y : Int
  , resource_a : Int
  , resource_b : Int
  , resource_c : Int
  , available_steps : Int
}

--state
type alias Model =
  { player : Player
  , error : String
  }

player_decoder : Decode.Decoder Player
player_decoder = 
  let
    map7 = Decode.map7
    field = Decode.field
    string = Decode.string
    int = Decode.int
  in
    map7 Player
      (field "Name" string)
      (field "Pos_x" int)
      (field "Pos_y" int)
      (field "Resource_A" int)
      (field "Resource_B" int)
      (field "Resource_C" int)
      (field "Available_steps" int)

load_player_data: Cmd Msg
load_player_data = Http.send PlayerDataArrived (Http.get "/get_data" player_decoder)

init: (Model, Cmd Msg)
init = (Model (Player "" 0 0 0 0 0 0) "", load_player_data)

--names of things that can happen
type Msg = LoadPlayerData | PlayerDataArrived (Result Http.Error Player)

nav_bar: Html Msg
nav_bar =
  nav [class "navbar navbar-inverse"]
  [ div [class "container-fluid"]
    [ div [class "navbar-header"]
      [ a [class "navbar-brand", href "/home"] [text "Browser Game"]
      ]
    , ul [class "nav navbar-nav"]
      [ li [class "active"] [a [href "/home"] [text "Home"]]
      , li [] [a [href "/world"] [text "World"]]
      ]
    , ul [class "nav navbar-nav navbar-right"]
      [ li [] [a [href "/unlogin"] [span [class "glyphicon glyphicon-log-out"] [], text " Unlogin"]]
      ]
    ]
  ]

info_table: Model -> Html Msg
info_table model =
  table [class "table table-condensed"]
        [ thead []
          [ tr []
            [ th [] [text "name"]
            , th [] [text "pos x"]
            , th [] [text "pos y"]
            , th [] [text "resource A"]
            , th [] [text "resource B"]
            , th [] [text "resource C"]
            , th [] [text "available steps"]
            ]
          ]
        , tbody []
          [ tr []
            [ td [] [model.player.name |> text]
            , td [] [model.player.pos_x |> toString |> text]
            , td [] [model.player.pos_y |> toString |> text]
            , td [] [model.player.resource_a |> toString |> text]
            , td [] [model.player.resource_b |> toString |> text]
            , td [] [model.player.resource_c |> toString |> text]
            , td [] [model.player.available_steps |> toString |> text]
            ]
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
    [ info_table model
    , button [onClick LoadPlayerData] [text "Load Data"]
    , error_message model
    ]
  ]

--what to do if a thing happens
update: Msg -> Model -> (Model, Cmd Msg)
update msg model =
    case msg of
        LoadPlayerData -> (model, load_player_data)
        PlayerDataArrived (Ok player) -> ({model | player=player}, Cmd.none)
        PlayerDataArrived (Err err) -> case err of 
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
