;; This Source Code Form is subject to the terms of the Mozilla Public
;; License, v. 2.0. If a copy of the MPL was not distributed with this
;; file, You can obtain one at https://mozilla.org/MPL/2.0/.

(ns cse-client.guide
  (:require [re-frame.core :as rf]
            [cse-client.controller :as controller]))

(def about-content [:div
                    [:p "This application demonstrates the features of the Core Simulation Environment, with basic
                    co-simulation functionality exposed through a simple graphical user interface. From this application
                    you will be able to:"]
                    [:ul
                     [:li "Load a configuration for co-simulation"]
                     [:li "Apply basic simulation control"]
                     [:li "Observe and trend any simulation variables"]
                     [:li "Log simulation variables to file"]]
                    [:p "The application is intended to be a generic tool for running simple co-simulations with the
                    cse-core simulation engine."]])

(def config-content [:div
                     [:h4 "Creating a co-simulation"]
                     [:p "The co-simulation configuration should be specified in the 'Configuration' input field in the
                     simulation setup page, and should consist of either:"
                      [:ul
                       [:li "the path to a directory holding one or more FMUs."]
                       [:li "the path to an 'OspSystemStructure.xml' file, or a directory containing one."]
                       [:li "the path to a 'SystemStructure.ssd' file, or a directory containing one."]]
                      [:p "Additionally, a path for storage of output log files can be specified in the 'Log output'
                      field. If this is not specified, no log files will be created."]]

                     [:h4 "Configuration with an 'OspSystemStructure.xml' file"]
                     [:p "A simulation can be set up using the CSE configuration format, which supports the OSP MSMI
                     standard. The configuration file describes which FMUs are part of the simulation and the
                     connections between them. Also in support of the standard, but not covered here, each FMU may be
                     bundled with an extended model description file, describing groups of input variables (sockets) and
                     output variables (plugs), as well as bonds (sets of plugs and sockets).
                     An example file can be downloaded below."]
                     [:a {:target "TOP" :href "static/xml/OspSystemStructure.xml" :download ""}
                      "Example OspSystemStructure.xml file"]

                     [:h4 "Configuration with a 'SystemStructure.ssd' file"]
                     [:p "A simulation can be set up using the SSP standard for FMUs. The SSP configuration file
                     describes which FMUs are part of the simulation and the connections between them.
                     An example SSP file can be downloaded below."]
                     [:a {:target "TOP" :href "static/xml/SystemStructure.ssd" :download ""}
                      "Example SystemStructure.ssd file"]

                     [:h4 "Creating a log configuration file"]
                     [:p "For configuring what signals to log in each FMU, cse-core optionally supports the inclusion of
                      a log config XML file, to be put
                     in the same directory as the FMUs."]
                     [:i "The file must be named LogConfig.xml (including case)."]
                     [:p ""]
                     [:p "An example file is provided below."]
                     [:p "Note that while a config can be provided in one file for all the simulators in the simulation,
                      output will still be to one file per simulator / FMU. An optional attribute"
                      [:i " decimationFactor"] " can be provided on the simulator level, which will specify how often
                     simulator values are to be logged (logging only every decimationFactor sample)."]
                     [:p "Finally, to log all signals for a simulator, simply leave the variable list empty as shown in
                     the example file."]
                     [:a {:target "TOP" :href "static/xml/LogConfig.xml" :download ""}
                      "Example logger configuration file"]

                     [:h4 "Note on JVM-based FMUs"]
                     [:p [:div "Loading FMUs which internally load a Java Virtual Machine will cause this demo
                     application to crash without warning. A viable workaround is to load the offending FMU using "
                          [:a {:href "https://github.com/NTNU-IHB/FMU-proxy"} "FMU-proxy."]]]])

(def simulation-content [:div
                         [:h4 "Running a simulation"]
                         [:p "Once the simulation has been loaded successfully, the instantiated FMUs (simulators) will
                         be listed in the left sidebar. By clicking their names, the individual simulators and their
                         variables can be explored. To start the simulation, click the \"Play\" button in the top
                         right."]

                         [:h4 "Modifying and plotting variables"]
                         [:p "To override a variable value, navigate to the variable you want to override and click the
                         \"edit\" icon in the variable table."]
                         [:p "To plot variables, first click \"Create new time series\" or  \"Create new XY plot\".
                          To include a variable in a plot, navigate to the variable and click the \"Add to plot\"
                          menu."]

                         [:h4 "Running scenarios"]
                         [:p "A scenario describes a time-based sequence of events, where each event modifies a variable
                         in a certain way. All events will be executed at simulation time = event time + simulation time
                          when the scenario was started. When a scenario is finished, either by reaching the specified
                          end time or by being stopped by the user, all variables which were modified during the course
                          of the scenario will be reset to their original values."]
                         [:p "Scenario files are served from a 'scenarios' subdirectory of the loaded configuration
                         directory."]
                         [:a {:target "TOP" :href "static/xml/scenario1.json" :download ""}
                          "Example scenario file"]])

(def guide-data (array-map "About" about-content
                           "Configuration" config-content
                           "Simulation" simulation-content))

(defn form []
  (let [active @(rf/subscribe [:active-guide-tab])]
    [:div.ui.one.column.grid
     [:div.one.column.row
      [:div.column
       [:div.ui.top.attached.tabular.menu
        (map (fn [[header _]]
               ^{:key (str "guide-tab-" header)}
               [:a.item {:data-tab header
                         :class    (when (= header active) "active")
                         :on-click #(rf/dispatch [::controller/guide-navigate header])}
                header]) guide-data)]
       (map (fn [[header content]]
              ^{:key (str "guide-content-" header)}
              [:div.ui.bottom.attached.tab.segment {:data-tab header
                                                    :class    (when (= header active) "active")}
               content]) guide-data)]]]))