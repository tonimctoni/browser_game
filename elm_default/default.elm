import Html exposing (Html, div, text, program, button)
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
init: (Model, Cmd Msg)
init = (Model "" 0 0 0 0 0 0 "", Cmd.none)

--names of things that can happen
type Msg = BtnClick | HttpDataArrived (Result Http.Error String)

--my_functions
--get_http_data : Cmd Msg
--get_http_data = Http.send HttpDataArrived (Http.get "http://localhost:8000/get_data" (Decode.at ["Name"] Decode.string))
--get_http_data = Http.send HttpDataArrived (Http.getString "http://localhost:8000/get_data")

--how it looks
view: Model -> Html Msg
view model =
    div []
        [button [onClick BtnClick] [text "Add A"]
        , Html.p [] [text model.name]
        , Html.p [] [model.pos_x |> toString |> text, ", " |> text, model.pos_y |> toString |> text]
        , Html.p [] [model.pos_y |> toString |> text]
        , Html.p [] [model.error |> text]
        ]

--what to do if a thing happens
update: Msg -> Model -> (Model, Cmd Msg)
update msg model =
    case msg of
        BtnClick -> (model, Http.send HttpDataArrived (Http.getString "http://localhost:8000/get_data"))
        HttpDataArrived (Ok json_string) -> 
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
                resource_a = case Decode.decodeString (Decode.field "Resource_a" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                resource_b = case Decode.decodeString (Decode.field "Resource_b" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                resource_c = case Decode.decodeString (Decode.field "Resource_c" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
                available_steps = case Decode.decodeString (Decode.field "Available_steps" Decode.int) json_string of
                    Ok(x) -> x
                    Err(_) -> -1
            in
                ({model | name=name
                    , pos_x=pos_x
                    , pos_y=pos_y
                    , resource_a=resource_a
                    , resource_b=resource_b
                    , resource_c=resource_c
                    , available_steps=available_steps
                    }, Cmd.none)
        HttpDataArrived (Err err) -> case err of 
            Http.BadUrl s -> ({model | error="BadUrl: "++s}, Cmd.none)
            Http.Timeout -> ({model | error="Timeout"}, Cmd.none)
            Http.NetworkError -> ({model | error="NetworkError"}, Cmd.none)
            Http.BadStatus _ -> ({model | error="BadStatus"}, Cmd.none)
            Http.BadPayload _ _ -> ({model | error="BadPayload"}, Cmd.none)

--("Error", Cmd.none)
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
