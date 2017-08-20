import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode

--state
type alias Model =
  { name : String
  , pos_x : Int
  , pos_y : Int
  , resource_a : Int
  , resource_b : Int
  , resource_c : Int
  , available_steps : Int
  , error : String
  }

load_player_data: Cmd Msg
load_player_data = Http.send PlayerDataArrived (Http.getString "/get_data")

init: (Model, Cmd Msg)
init = (Model "" 0 0 0 0 0 0 "", load_player_data)

--names of things that can happen
type Msg = LoadPlayerData | PlayerDataArrived (Result Http.Error String)

nav_bar: Html Msg
nav_bar =
  nav [class "navbar navbar-inverse"]
  [ div [class "container-fluid"]
    [ div [class "navbar-header"]
      [ a [class "navbar-brand", href "/home"] [text "Browser Game"]
      ]
    , ul [class "nav navbar-nav"]
      [ li [class "active"] [a [href "/#"] [text "Home"]]
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
            [ td [] [model.name |> text]
            , td [] [model.pos_x |> toString |> text]
            , td [] [model.pos_y |> toString |> text]
            , td [] [model.resource_a |> toString |> text]
            , td [] [model.resource_b |> toString |> text]
            , td [] [model.resource_c |> toString |> text]
            , td [] [model.available_steps |> toString |> text]
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
        PlayerDataArrived (Ok json_string) -> 
            let
                name = case Decode.decodeString (Decode.field "Name" Decode.string) json_string of
                    Ok(x) -> x
                    Err(_) -> "error"
                pos_x = case Decode.decodeString (Decode.field "Pos_x" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                pos_y = case Decode.decodeString (Decode.field "Pos_y" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                resource_a = case Decode.decodeString (Decode.field "Resource_A" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                resource_b = case Decode.decodeString (Decode.field "Resource_B" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                resource_c = case Decode.decodeString (Decode.field "Resource_C" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                available_steps = case Decode.decodeString (Decode.field "Available_steps" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                error = ""
            in
                ({model | name=name
                    , pos_x=pos_x
                    , pos_y=pos_y
                    , resource_a=resource_a
                    , resource_b=resource_b
                    , resource_c=resource_c
                    , available_steps=available_steps
                    , error=error
                    }, Cmd.none)
        PlayerDataArrived (Err err) -> case err of 
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
