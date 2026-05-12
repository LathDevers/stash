import React, { PropsWithChildren, useEffect } from "react";
import { CollapseButton } from "./CollapseButton";
import { useOnOutsideClick } from "src/hooks/OutsideClick";
import ScreenUtils, { useMediaQuery } from "src/utils/screen";
import { View } from "../List/views";
import cx from "classnames";
import { Button, CollapseProps } from "react-bootstrap";
import { useIntl } from "react-intl";
import { Icon } from "./Icon";
import { faSliders } from "@fortawesome/free-solid-svg-icons";

export type SidebarSectionStates = Record<string, boolean>;

// this needs to correspond to the CSS media query that overlaps the sidebar over content
const fixedSidebarMediaQuery = "only screen and (max-width: 767px)";

export const Sidebar: React.FC<
  PropsWithChildren<{
    hide?: boolean;
    onHide?: () => void;
  }>
> = ({ hide, onHide, children }) => {
  const ref = React.useRef<HTMLDivElement>(null);

  const closeOnOutsideClick = useMediaQuery(fixedSidebarMediaQuery) && !hide;

  useOnOutsideClick(
    ref,
    !closeOnOutsideClick ? undefined : onHide,
    "ignore-sidebar-outside-click"
  );

  return (
    <div ref={ref} className="sidebar">
      {children}
    </div>
  );
};

// SidebarPane is a container for a Sidebar and content.
// It is expected that the children will be two elements:
// a Sidebar and a content element.
export const SidebarPane: React.FC<
  PropsWithChildren<{
    hideSidebar?: boolean;
  }>
> = ({ hideSidebar = false, children }) => {
  return (
    <div className={cx("sidebar-pane", { "hide-sidebar": hideSidebar })}>
      {children}
    </div>
  );
};

export const SidebarToggleButton: React.FC<{
  onClick: () => void;
}> = ({ onClick }) => {
  const intl = useIntl();
  return (
    <div className="sidebar-toggle-button-container">
      <Button
        className="sidebar-toggle-button ignore-sidebar-outside-click minimal"
        variant="secondary"
        onClick={onClick}
        title={intl.formatMessage({ id: "actions.sidebar.toggle" })}
      >
        <Icon icon={faSliders} />
      </Button>
    </div>
  );
};

export const SidebarPaneContent: React.FC<{ onSidebarToggle: () => void }> = ({
  onSidebarToggle,
  children,
}) => {
  return (
    <div className="sidebar-pane-content">
      <SidebarToggleButton onClick={onSidebarToggle} />
      {children}
    </div>
  );
};

interface IContext {
  sectionOpen: SidebarSectionStates;
  setSectionOpen: (section: string, open: boolean) => void;
}

export const SidebarStateContext = React.createContext<IContext | null>(null);

export interface ISidebarSectionProps {
  text: React.ReactNode;
  className?: string;
  outsideCollapse?: React.ReactNode;
  onOpen?: () => void;
  // used to store open/closed state in SidebarStateContext
  sectionID?: string;
}

export const SidebarSection: React.FC<
  PropsWithChildren<ISidebarSectionProps>
> = ({
  className = "",
  text,
  outsideCollapse,
  onOpen,
  sectionID = "",
  children,
}) => {
  // this is optional
  const contextState = React.useContext(SidebarStateContext);
  const openState =
    !contextState || !sectionID
      ? undefined
      : contextState.sectionOpen[sectionID] ?? undefined;

  function onOpenInternal(open: boolean) {
    if (contextState && sectionID) {
      contextState.setSectionOpen(sectionID, open);
    }
  }

  useEffect(() => {
    if (openState && onOpen) {
      onOpen();
    }
  }, [openState, onOpen]);

  const collapseProps: Partial<CollapseProps> = {
    mountOnEnter: true,
    unmountOnExit: true,
  };
  return (
    <CollapseButton
      className={`sidebar-section ${className}`}
      collapseProps={collapseProps}
      text={text}
      outsideCollapse={outsideCollapse}
      onOpenChanged={onOpenInternal}
      open={openState}
    >
      {children}
    </CollapseButton>
  );
};

// show sidebar by default if not on mobile
export function defaultShowSidebar() {
  return !ScreenUtils.matchesMediaQuery(fixedSidebarMediaQuery);
}

// Filter sidebar is currently dropped — the Edit Filter dialog covers the
// same surface and the sidebar is reserved for future "categories" navigation.
// To re-enable, restore the previous body of this hook from git history.
export function useSidebarState(_view?: View) {
  return {
    showSidebar: false,
    sectionOpen: {} as SidebarSectionStates,
    setShowSidebar: (_show: boolean | ((prev: boolean | undefined) => boolean)) =>
      undefined,
    setSectionOpen: (_section: string, _open: boolean) => undefined,
    loading: false,
  };
}
