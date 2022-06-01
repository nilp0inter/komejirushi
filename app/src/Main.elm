port module Main exposing (main)

import Browser
import Debug
import Dict as D
import List as L
import Html exposing (Html)
import Html.Attributes as HA
import Html.Events as HE
import Json.Encode as JE
import Json.Decode as JD

-- JavaScript usage: app.ports.websocketIn.send(response);
port websocketIn : (String -> msg) -> Sub msg
-- JavaScript usage: app.ports.websocketOut.subscribe(handler);
port websocketOut : String -> Cmd msg

main = Browser.element
    { init = init
    , update = update
    , view = view
    , subscriptions = subscriptions
    }

{- MODEL -}

type alias Model =
    { responses : List WSMsg
    , input : String
    , results : List Entry
    }

init : () -> (Model, Cmd Msg)
init _ =
    ( { responses = []
      , input = ""
      , results = []
      }
    , Cmd.none
    )

type alias Entry =
    { name : String
    , score : Int
    , url : String
    , docset : String
    }

type alias SResult =
    { name : String
    , score : Int
    , url : String
    }

type alias SResults = 
    { results : D.Dict String (List SResult) }

type WSMsg = SearchResult SResults

type Msg = Change String
    | Submit String
    | WebsocketIn WSMsg
    | WebsocketError JD.Error

{- DECODER -}

sResultDecoder : JD.Decoder SResult
sResultDecoder = JD.map3 SResult
  (JD.field "n" JD.string)
  (JD.field "s" JD.int)
  (JD.field "u" JD.string)

wsMsgDecoder : JD.Decoder WSMsg
wsMsgDecoder = JD.map SearchResult (JD.map SResults (JD.field "results" (JD.dict (JD.list sResultDecoder))))

wsDecoder : String -> Msg
wsDecoder s = case (JD.decodeString wsMsgDecoder s) of
  Ok v -> WebsocketIn v
  Err e -> WebsocketError e

toEntryList : WSMsg -> List Entry
toEntryList s = case s of
  SearchResult rs -> D.foldr (\k v a -> (L.map (\e -> { name = .name e, score = .score e, url = .url e, docset = k }) v) ++ a) [] (.results rs)

{- UPDATE -}

update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    Change input ->
      ( { model | input = input }
      , Cmd.none
      )
    Submit value ->
      ( { model | results = [] }
      , websocketOut value
      )
    WebsocketIn m ->
      ( { model | results = L.sortWith (\a b -> compare (.score b) (.score a)) (model.results ++ toEntryList m)}
      , Cmd.none
      )
    WebsocketError err ->
      ( { model | results = [] }
      , Cmd.none
      )

{- SUBSCRIPTIONS -}

subscriptions : Model -> Sub Msg
subscriptions model =
    websocketIn wsDecoder

{- VIEW -}

li : String -> Html Msg
li string = Html.li [] [Html.text string]

view : Model -> Html Msg
view model = Html.div []
    --[ Html.form [HE.onSubmit (WebsocketIn model.input)] -- Short circuit to test without ports
    [ Html.form [HE.onSubmit (Submit model.input)]
      [ Html.input [HA.placeholder "Enter some text.", HA.value model.input, HE.onInput Change] []
      , model.results |> L.take 10 |> L.map (\e -> Html.li [] [Html.a [ HA.href (.url e) ] [Html.text (.name e)]]) |> Html.ol []
      ]
    ]
    
